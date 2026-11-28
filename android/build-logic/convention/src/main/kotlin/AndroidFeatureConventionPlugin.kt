import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.kotlin.dsl.dependencies

class AndroidFeatureConventionPlugin : Plugin<Project> {
    override fun apply(target: Project) = with(target) {
        pluginManager.apply("trivia.android.library")
        pluginManager.apply("trivia.android.compose")
        pluginManager.apply("trivia.hilt")
        dependencies {
            add("implementation", project(":core:ui"))
            add("implementation", project(":core:designsystem"))
            add("implementation", project(":core:network"))
            add("implementation", lib("androidx-hilt-navigation-compose"))
        }
    }
}
