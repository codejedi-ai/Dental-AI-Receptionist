import SwiftUI

struct DialerView: View {

    @State private var displayNumber = ""
    @State private var navigateToCall = false

    private let keyRows = [["1","2","3"], ["4","5","6"], ["7","8","9"], ["*","0","#"]]

    var body: some View {
        NavigationStack {
            ZStack {
                backgroundGradient.ignoresSafeArea()

                VStack(spacing: 20) {

                    // Title
                    Text("Smile Dental")
                        .font(.system(size: 18, weight: .medium))
                        .foregroundStyle(Color(hex: "#64B5F6"))
                        .padding(.top, 12)

                    // Number display
                    HStack {
                        Spacer()
                        Text(displayNumber.isEmpty ? " " : displayNumber)
                            .font(.system(size: 36, weight: .light, design: .monospaced))
                            .foregroundStyle(.white)
                            .lineLimit(1)
                            .minimumScaleFactor(0.4)
                            .padding(.trailing, 4)
                    }
                    .padding(.horizontal, 20)
                    .padding(.vertical, 16)
                    .background(Color(hex: "#16213e"))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .padding(.horizontal, 24)

                    // Keypad
                    VStack(spacing: 10) {
                        ForEach(keyRows, id: \.self) { row in
                            HStack(spacing: 10) {
                                ForEach(row, id: \.self) { key in
                                    KeypadButton(label: key) { appendKey(key) }
                                }
                            }
                        }
                    }
                    .padding(.horizontal, 32)

                    // Bottom row: backspace + call
                    HStack(spacing: 48) {
                        // Backspace
                        Button {
                            backspace()
                        } label: {
                            Image(systemName: "delete.left")
                                .font(.system(size: 20, weight: .medium))
                                .foregroundStyle(.white)
                                .frame(width: 60, height: 60)
                                .background(Color(hex: "#607D8B"))
                                .clipShape(Circle())
                        }
                        .simultaneousGesture(LongPressGesture(minimumDuration: 0.6).onEnded { _ in
                            displayNumber = ""
                        })

                        // Call button
                        Button {
                            if !displayNumber.isEmpty { navigateToCall = true }
                        } label: {
                            Image(systemName: "phone.fill")
                                .font(.system(size: 24, weight: .medium))
                                .foregroundStyle(.white)
                                .frame(width: 72, height: 72)
                                .background(displayNumber.isEmpty ? Color.gray.opacity(0.6) : Color(hex: "#4CAF50"))
                                .clipShape(Circle())
                        }
                        .disabled(displayNumber.isEmpty)
                    }
                    .padding(.top, 4)

                    // Recent shortcuts
                    if displayNumber.isEmpty {
                        quickDials
                    }

                    Spacer()
                }
            }
            .navigationDestination(isPresented: $navigateToCall) {
                CallView(number: displayNumber)
            }
        }
    }

    // MARK: - Sub-views

    private var backgroundGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "#1a1a2e"), Color(hex: "#0f0f1a")],
            startPoint: .top, endPoint: .bottom
        )
    }

    private var quickDials: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Quick Dial")
                .font(.caption)
                .foregroundStyle(.gray)
                .padding(.horizontal, 28)

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 12) {
                    ForEach([("101","Receptionist"),("102","Billing"),("103","Reminders")], id: \.0) { number, name in
                        Button {
                            displayNumber = number
                        } label: {
                            VStack(spacing: 4) {
                                Text(number)
                                    .font(.system(size: 16, weight: .semibold, design: .monospaced))
                                    .foregroundStyle(.white)
                                Text(name)
                                    .font(.caption2)
                                    .foregroundStyle(.gray)
                            }
                            .padding(.horizontal, 16)
                            .padding(.vertical, 10)
                            .background(Color(hex: "#2c2c3e"))
                            .clipShape(RoundedRectangle(cornerRadius: 10))
                        }
                    }
                }
                .padding(.horizontal, 24)
            }
        }
        .padding(.top, 8)
    }

    // MARK: - Actions

    private func appendKey(_ key: String) {
        guard displayNumber.count < 15 else { return }
        displayNumber += key
    }

    private func backspace() {
        guard !displayNumber.isEmpty else { return }
        displayNumber.removeLast()
    }
}

// MARK: - KeypadButton

struct KeypadButton: View {
    let label: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(label)
                .font(.system(size: 24, weight: .medium))
                .foregroundStyle(.white)
                .frame(width: 76, height: 76)
                .background(Color(hex: "#2c2c3e"))
                .clipShape(Circle())
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Color helper

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3:  (a,r,g,b) = (255, (int>>8)*17, (int>>4 & 0xF)*17, (int & 0xF)*17)
        case 6:  (a,r,g,b) = (255, int>>16, int>>8 & 0xFF, int & 0xFF)
        case 8:  (a,r,g,b) = (int>>24, int>>16 & 0xFF, int>>8 & 0xFF, int & 0xFF)
        default: (a,r,g,b) = (255, 0, 0, 0)
        }
        self.init(.sRGB, red: Double(r)/255, green: Double(g)/255, blue: Double(b)/255, opacity: Double(a)/255)
    }
}

#Preview {
    DialerView()
}
