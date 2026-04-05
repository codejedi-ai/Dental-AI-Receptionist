import Foundation

@MainActor
protocol WebSocketManagerDelegate: AnyObject {
    func onRinging(agentName: String)
    func onConnected()
    func onBusy(reason: String)
    func onEnded(reason: String)
    func onAudioReceived(_ data: Data)
    func onError(_ message: String)
}

/// Manages the WebSocket connection to the backend server.
/// Sends the dial message on open, forwards audio, and dispatches control events to the delegate.
class WebSocketManager: NSObject {

    weak var delegate: (any WebSocketManagerDelegate)?

    private var webSocketTask: URLSessionWebSocketTask?
    private lazy var urlSession = URLSession(
        configuration: .default,
        delegate: self,
        delegateQueue: OperationQueue.main   // delegate callbacks on main thread
    )
    private var isOpen = false
    private var pendingNumber: String?

    // MARK: - Public API

    func connect(to urlString: String, dialingNumber number: String) {
        guard let url = URL(string: urlString) else {
            Task { @MainActor [weak self] in self?.delegate?.onError("Invalid server URL: \(urlString)") }
            return
        }
        pendingNumber = number
        webSocketTask = urlSession.webSocketTask(with: url)
        webSocketTask?.resume()
        receiveNext()
    }

    func sendAudio(_ data: Data) {
        guard isOpen else { return }
        webSocketTask?.send(.data(data)) { _ in }
    }

    func sendHangup() {
        sendJSON(["action": "hangup"])
    }

    func sendMute(_ muted: Bool) {
        sendJSON(["action": "mute", "muted": muted] as [String: Any])
    }

    func disconnect() {
        isOpen = false
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
    }

    // MARK: - Private

    private func sendDial(number: String) {
        sendJSON(["action": "dial", "number": number])
    }

    private func sendJSON(_ dict: [String: Any]) {
        guard let data = try? JSONSerialization.data(withJSONObject: dict),
              let text = String(data: data, encoding: .utf8) else { return }
        webSocketTask?.send(.string(text)) { _ in }
    }

    private func receiveNext() {
        webSocketTask?.receive { [weak self] result in
            // Completion runs on URLSession's internal queue — hop to main actor
            Task { @MainActor [weak self] in
                guard let self else { return }
                switch result {
                case .success(let message):
                    self.handle(message)
                    self.receiveNext()
                case .failure(let error):
                    self.delegate?.onError(error.localizedDescription)
                }
            }
        }
    }

    private func handle(_ message: URLSessionWebSocketTask.Message) {
        switch message {
        case .data(let data):
            delegate?.onAudioReceived(data)

        case .string(let text):
            guard let jsonData = text.data(using: .utf8),
                  let json = try? JSONSerialization.jsonObject(with: jsonData) as? [String: Any],
                  let event = json["event"] as? String else { return }
            switch event {
            case "ringing":
                delegate?.onRinging(agentName: json["agentName"] as? String ?? "")
            case "connected":
                delegate?.onConnected()
            case "busy":
                delegate?.onBusy(reason: json["reason"] as? String ?? "Unavailable")
            case "ended":
                delegate?.onEnded(reason: json["reason"] as? String ?? "ended")
            default:
                break
            }

        @unknown default:
            break
        }
    }
}

// MARK: - URLSessionWebSocketDelegate
extension WebSocketManager: URLSessionWebSocketDelegate {

    nonisolated func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didOpenWithProtocol protocol: String?
    ) {
        // Called on OperationQueue.main because that's the delegate queue
        Task { @MainActor [weak self] in
            guard let self else { return }
            self.isOpen = true
            if let number = self.pendingNumber {
                self.sendDial(number: number)
            }
        }
    }

    nonisolated func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didCloseWith closeCode: URLSessionWebSocketTask.CloseCode,
        reason: Data?
    ) {
        Task { @MainActor [weak self] in
            guard let self else { return }
            self.isOpen = false
            self.delegate?.onEnded(reason: "connection_closed")
        }
    }
}
