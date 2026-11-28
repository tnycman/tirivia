import com.android.build.api.dsl.ApplicationExtension
import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.kotlin.dsl.configure

// AGP 9 compiles Kotlin itself; kotlin-android is never applied (research 14).
class AndroidApplicationConventionPlugin : Plugin<Project> {
    override fun apply(target: Project) = with(target) {
        pluginManager.apply("com.android.application")
        extensions.configure<ApplicationExtension> {
            compileSdk = sdk("compileSdk")
            defaultConfig {
                minSdk = sdk("minSdk")
                targetSdk = sdk("targetSdk")
            }
            androidResources { generateLocaleConfig = true }
        }
    }
}
