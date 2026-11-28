import DesignSystem
import SwiftUI

// Shell screen until the sign-in flow lands (PLAN Phase 1).
@main
struct TiriviaApp: App {
    var body: some Scene {
        WindowGroup {
            ZStack {
                Color(uiColor: TokensLight.themeBg).ignoresSafeArea()
                Text("Tirivia").foregroundStyle(Color(uiColor: TokensLight.themePrimaryText))
            }
        }
    }
}
