import Foundation

enum Config {
    /// Change this to your Mac/server's local IP address.
    /// Find it with: ifconfig | grep "inet 192"
    static let backendWSURL = "ws://192.168.1.100:3000/call"

    static let dialingTimeoutSeconds: TimeInterval = 30
    static let ringingTimeoutSeconds: TimeInterval = 60
    static let endedAutoReturnSeconds: TimeInterval = 5
}
