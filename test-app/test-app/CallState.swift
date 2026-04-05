import Foundation

enum CallState: Equatable {
    case dialing
    case ringing(agentName: String)
    case connected
    case ended(reason: String)

    var isTerminal: Bool {
        if case .ended = self { return true }
        return false
    }
}
