import AVFoundation

/// Handles real-time audio capture (microphone → PCM 16kHz Int16) and playback
/// (received PCM 16kHz Int16 → speaker).  All playback is scheduled on the
/// AVAudioPlayerNode; capture is tapped from the input node and converted to
/// the wire format before being handed to `onAudioCaptured`.
class AudioEngine {

    // Wire format: PCM 16-bit, mono, 16 kHz — matches protocol spec
    private let wireFormat: AVAudioFormat = {
        AVAudioFormat(commonFormat: .pcmFormatInt16, sampleRate: 16_000, channels: 1, interleaved: true)!
    }()

    private let engine = AVAudioEngine()
    private let playerNode = AVAudioPlayerNode()

    // Accessed from the tap callback (background thread) — marked nonisolated(unsafe)
    nonisolated(unsafe) var onAudioCaptured: ((Data) -> Void)?
    nonisolated(unsafe) var isMuted = false

    private var captureConverter: AVAudioConverter?

    // MARK: - Start / Stop

    func start() throws {
        let session = AVAudioSession.sharedInstance()
        try session.setCategory(.playAndRecord, mode: .voiceChat, options: .defaultToSpeaker)
        try session.setActive(true)

        // Connect playerNode → mainMixerNode using wire format.
        // AVAudioEngine inserts an auto-format-converter as needed.
        engine.attach(playerNode)
        engine.connect(playerNode, to: engine.mainMixerNode, format: wireFormat)

        // Build a converter from the hardware input format → wire format
        let inputNode = engine.inputNode
        let inputFormat = inputNode.outputFormat(forBus: 0)
        captureConverter = AVAudioConverter(from: inputFormat, to: wireFormat)

        // Install tap at the hardware input format
        inputNode.installTap(onBus: 0, bufferSize: 1_024, format: inputFormat) { [weak self] buffer, _ in
            self?.processCapture(buffer)
        }

        try engine.start()
        playerNode.play()
    }

    func stop() {
        engine.inputNode.removeTap(onBus: 0)
        playerNode.stop()
        engine.stop()
        try? AVAudioSession.sharedInstance().setActive(false, options: .notifyOthersOnDeactivation)
    }

    // MARK: - Playback

    func playReceived(_ data: Data) {
        let frameCount = data.count / 2
        guard frameCount > 0,
              let buffer = AVAudioPCMBuffer(pcmFormat: wireFormat, frameCapacity: AVAudioFrameCount(frameCount))
        else { return }

        buffer.frameLength = AVAudioFrameCount(frameCount)
        data.withUnsafeBytes { raw in
            guard let src = raw.baseAddress?.assumingMemoryBound(to: Int16.self),
                  let dst = buffer.int16ChannelData else { return }
            dst[0].update(from: src, count: frameCount)
        }
        playerNode.scheduleBuffer(buffer)
    }

    // MARK: - Private capture

    // Called on AVAudioEngine's internal (background) thread.
    private func processCapture(_ buffer: AVAudioPCMBuffer) {
        guard !isMuted, let converter = captureConverter else { return }

        let inputRate  = buffer.format.sampleRate
        let outputRate = wireFormat.sampleRate
        let outputCapacity = AVAudioFrameCount(Double(buffer.frameLength) * outputRate / inputRate) + 1

        guard let outBuffer = AVAudioPCMBuffer(pcmFormat: wireFormat, frameCapacity: outputCapacity) else { return }

        var inputConsumed = false
        var error: NSError?
        converter.convert(to: outBuffer, error: &error) { _, outStatus in
            if inputConsumed { outStatus.pointee = .noDataNow; return nil }
            outStatus.pointee  = .haveData
            inputConsumed = true
            return buffer
        }

        guard error == nil, outBuffer.frameLength > 0,
              let channelData = outBuffer.int16ChannelData else { return }

        let data = Data(bytes: channelData[0], count: Int(outBuffer.frameLength) * 2)
        onAudioCaptured?(data)
    }
}
