import SwiftUI
import AVFoundation

// MARK: - ViewModel

@MainActor
@Observable
class CallViewModel: NSObject, WebSocketManagerDelegate {

    var state: CallState = .dialing
    var agentName: String = ""
    var callDurationSeconds: Int = 0
    var isMuted = false
    var waveformAmplitudes: [CGFloat] = Array(repeating: 0.2, count: 7)

    let number: String

    private let wsManager = WebSocketManager()
    private var audioEngine: AudioEngine?
    private var callTimer: Timer?
    private var timeoutTimer: Timer?

    init(number: String) {
        self.number = number
        super.init()
        wsManager.delegate = self
    }

    // MARK: - Call lifecycle

    func startCall() {
        wsManager.connect(to: Config.backendWSURL, dialingNumber: number)
        scheduleTimeout(Config.dialingTimeoutSeconds, reason: "Connection timeout")
    }

    func hangUp() {
        wsManager.sendHangup()
        endCall(reason: "You ended the call")
    }

    func toggleMute() {
        isMuted.toggle()
        audioEngine?.isMuted = isMuted
        wsManager.sendMute(isMuted)
    }

    var durationFormatted: String {
        String(format: "%02d:%02d", callDurationSeconds / 60, callDurationSeconds % 60)
    }

    // MARK: - Private

    private func scheduleTimeout(_ interval: TimeInterval, reason: String) {
        timeoutTimer?.invalidate()
        timeoutTimer = Timer.scheduledTimer(withTimeInterval: interval, repeats: false) { [weak self] _ in
            Task { @MainActor [weak self] in
                guard let self, !self.state.isTerminal else { return }
                self.endCall(reason: reason)
            }
        }
    }

    private func startAudio() {
        let engine = AudioEngine()
        audioEngine = engine

        engine.onAudioCaptured = { [weak self] data in
            // Called from AVAudioEngine's background thread
            self?.wsManager.sendAudio(data)
            Task { @MainActor [weak self] in
                self?.updateWaveform(from: data)
            }
        }

        do {
            try engine.start()
        } catch {
            print("AudioEngine start failed: \(error)")
        }

        callTimer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in self?.callDurationSeconds += 1 }
        }
    }

    private func endCall(reason: String) {
        timeoutTimer?.invalidate()
        callTimer?.invalidate()
        audioEngine?.stop()
        wsManager.disconnect()
        state = .ended(reason: reason)
    }

    private func updateWaveform(from data: Data) {
        let samples = data.withUnsafeBytes { ptr -> [Int16] in
            Array(ptr.bindMemory(to: Int16.self))
        }
        guard !samples.isEmpty else { return }
        let barCount = waveformAmplitudes.count
        let step = max(1, samples.count / barCount)
        waveformAmplitudes = (0..<barCount).map { i in
            let slice = samples[(i * step)..<min((i + 1) * step, samples.count)]
            let avg = slice.map { abs(Double($0)) }.reduce(0, +) / Double(slice.count)
            return CGFloat(min(avg / 32_768.0, 1.0))
        }
    }

    // MARK: - WebSocketManagerDelegate

    func onRinging(agentName: String) {
        timeoutTimer?.invalidate()
        self.agentName = agentName
        state = .ringing(agentName: agentName)
        scheduleTimeout(Config.ringingTimeoutSeconds, reason: "No answer")
    }

    func onConnected() {
        timeoutTimer?.invalidate()
        state = .connected
        startAudio()
    }

    func onBusy(reason: String) {
        endCall(reason: reason)
    }

    func onEnded(reason: String) {
        guard !state.isTerminal else { return }
        endCall(reason: reason)
    }

    func onAudioReceived(_ data: Data) {
        audioEngine?.playReceived(data)
    }

    func onError(_ message: String) {
        guard !state.isTerminal else { return }
        endCall(reason: "Error: \(message)")
    }
}

// MARK: - CallView

struct CallView: View {

    let number: String
    @State private var viewModel: CallViewModel
    @Environment(\.dismiss) private var dismiss

    init(number: String) {
        self.number = number
        _viewModel = State(initialValue: CallViewModel(number: number))
    }

    var body: some View {
        ZStack {
            LinearGradient(colors: [Color(hex: "#1a1a2e"), Color(hex: "#0f0f1a")],
                           startPoint: .top, endPoint: .bottom)
                .ignoresSafeArea()

            VStack {
                Spacer()
                stateContent
                Spacer()
            }
        }
        .navigationBarBackButtonHidden(true)
        .onAppear { viewModel.startCall() }
        .onChange(of: viewModel.state) { _, newState in
            if case .ended = newState {
                Task {
                    try? await Task.sleep(for: .seconds(Config.endedAutoReturnSeconds))
                    dismiss()
                }
            }
        }
    }

    // MARK: - State views

    @ViewBuilder
    private var stateContent: some View {
        switch viewModel.state {
        case .dialing:
            dialingView
        case .ringing(let name):
            ringingView(name: name)
        case .connected:
            connectedView
        case .ended(let reason):
            endedView(reason: reason)
        }
    }

    private var dialingView: some View {
        VStack(spacing: 28) {
            PulsingPhoneIcon(color: Color(hex: "#4CAF50"))

            VStack(spacing: 8) {
                Text("Calling \(number)...")
                    .font(.title2.weight(.medium))
                    .foregroundStyle(.white)
            }

            cancelButton
        }
    }

    private func ringingView(name: String) -> some View {
        VStack(spacing: 28) {
            PulsingPhoneIcon(color: Color(hex: "#4CAF50"))

            VStack(spacing: 8) {
                Text("Ringing...")
                    .font(.title2.weight(.medium))
                    .foregroundStyle(.white)

                if !name.isEmpty {
                    Text(name)
                        .font(.subheadline)
                        .foregroundStyle(Color(hex: "#64B5F6"))
                }
            }

            cancelButton
        }
    }

    private var connectedView: some View {
        VStack(spacing: 24) {
            if !viewModel.agentName.isEmpty {
                Text(viewModel.agentName)
                    .font(.headline)
                    .foregroundStyle(Color(hex: "#64B5F6"))
            }

            Text(viewModel.durationFormatted)
                .font(.system(size: 52, weight: .thin, design: .monospaced))
                .foregroundStyle(.white)

            WaveformBars(amplitudes: viewModel.waveformAmplitudes)
                .frame(height: 56)
                .padding(.horizontal, 48)

            // Controls
            HStack(spacing: 48) {
                CircleButton(
                    icon: viewModel.isMuted ? "mic.slash.fill" : "mic.fill",
                    background: viewModel.isMuted ? Color.red.opacity(0.85) : Color(hex: "#2c2c3e"),
                    action: { viewModel.toggleMute() }
                )

                CircleButton(
                    icon: "speaker.wave.2.fill",
                    background: Color(hex: "#2c2c3e"),
                    action: {}         // VoiceChat mode handles routing automatically
                )
            }

            hangupButton
        }
    }

    private func endedView(reason: String) -> some View {
        VStack(spacing: 24) {
            Image(systemName: "phone.down.fill")
                .font(.system(size: 56))
                .foregroundStyle(.red)

            Text("Call Ended")
                .font(.title.weight(.medium))
                .foregroundStyle(.white)

            if viewModel.callDurationSeconds > 0 {
                Text("Duration: \(viewModel.durationFormatted)")
                    .font(.subheadline)
                    .foregroundStyle(.gray)
            }

            Text(reason)
                .font(.caption)
                .foregroundStyle(.gray)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 40)

            Button {
                dismiss()
            } label: {
                Text("Back to Dialer")
                    .font(.headline)
                    .foregroundStyle(.white)
                    .padding(.horizontal, 40)
                    .padding(.vertical, 14)
                    .background(Color(hex: "#607D8B"))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }
        }
    }

    // MARK: - Shared buttons

    private var cancelButton: some View {
        Button { viewModel.hangUp() } label: {
            Image(systemName: "phone.down.fill")
                .font(.system(size: 24))
                .foregroundStyle(.white)
                .frame(width: 72, height: 72)
                .background(.red)
                .clipShape(Circle())
        }
    }

    private var hangupButton: some View {
        Button { viewModel.hangUp() } label: {
            Label("Hang Up", systemImage: "phone.down.fill")
                .font(.headline)
                .foregroundStyle(.white)
                .padding(.horizontal, 40)
                .padding(.vertical, 16)
                .background(.red)
                .clipShape(RoundedRectangle(cornerRadius: 16))
        }
    }
}

// MARK: - Supporting Views

struct PulsingPhoneIcon: View {
    let color: Color
    @State private var pulsing = false

    var body: some View {
        Image(systemName: "phone.fill")
            .font(.system(size: 64))
            .foregroundStyle(color)
            .scaleEffect(pulsing ? 1.15 : 1.0)
            .opacity(pulsing ? 0.7 : 1.0)
            .animation(.easeInOut(duration: 0.9).repeatForever(autoreverses: true), value: pulsing)
            .onAppear { pulsing = true }
    }
}

struct WaveformBars: View {
    let amplitudes: [CGFloat]

    var body: some View {
        GeometryReader { geo in
            HStack(alignment: .center, spacing: 4) {
                ForEach(0..<amplitudes.count, id: \.self) { i in
                    RoundedRectangle(cornerRadius: 2)
                        .fill(Color(hex: "#4CAF50"))
                        .frame(width: (geo.size.width - CGFloat(amplitudes.count - 1) * 4) / CGFloat(amplitudes.count),
                               height: max(4, amplitudes[i] * geo.size.height))
                }
            }
            .frame(maxHeight: .infinity, alignment: .center)
        }
        .animation(.easeInOut(duration: 0.08), value: amplitudes)
    }
}

struct CircleButton: View {
    let icon: String
    let background: Color
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Image(systemName: icon)
                .font(.system(size: 22))
                .foregroundStyle(.white)
                .frame(width: 60, height: 60)
                .background(background)
                .clipShape(Circle())
        }
    }
}

#Preview {
    NavigationStack {
        CallView(number: "101")
    }
}
