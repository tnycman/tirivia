import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.kotlin.dsl.dependencies

class AndroidFeatureConventionPlugin : Plugin<Project> {
    override fun apply(target: Project) = with(target) {
        pluginManager.apply("tirivia.android.library")
        pluginManager.apply("tirivia.android.compose")
        pluginManager.apply("tirivia.hilt")
        dependencies {
            add("implementation", project(":core:ui"))
            add("implementation", project(":core:designsystem"))
            add("implementation", project(":core:network"))
            add("implementation", lib("androidx-hilt-navigation-compose"))
        }
    }
}
