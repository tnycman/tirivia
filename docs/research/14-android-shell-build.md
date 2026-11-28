# 14. Android shell: what it takes to compile

`android/` has the module graph, the version catalog and empty convention-plugin wiring, and it has never been compiled. It has no Gradle wrapper, no convention-plugin sources, no app source and no CI workflow. Sources read 2026-11-28 (verbatim unless marked).

## Build tools

| Source | Text |
|---|---|
| [AGP releases](https://developer.android.com/build/releases/gradle-plugin) | AGP 9.5: Gradle minimum 9.6.0, JDK minimum 17; "The maximum API level that Android Gradle plugin 9.5 supports is API level 37." |
| [AGP 9.0 notes](https://developer.android.com/build/releases/agp-9-0-0-release-notes) | "Android Gradle plugin 9.0 introduces built-in Kotlin support and enables it by default." "you no longer have to apply the `org.jetbrains.kotlin.android` ... plugin"; "The `org.jetbrains.kotlin.android` plugin is not compatible with the new DSL." |
| same | "AGP 9.0 uses our new DSL interfaces exclusively"; the block methods move to `ApplicationExtension`, `LibraryExtension`, ...; old `BaseExtension` casts fail with `ClassCastException`; `applicationVariants` etc. are removed (use `androidComponents`) |
| same | "AGP 9.0 has a runtime dependency on Kotlin Gradle plugin (KGP) 2.2.10." How to raise it to our Kotlin 2.4.21 is not stated on the pages read: UNVERIFIED, confirm in CI |
| [KSP releases](https://github.com/google/ksp/releases) | Latest 2.3.14. No published Kotlin-pairing statement for 2.3.x (UNVERIFIED); AGP only forces an upgrade below 2.2.10-2.0.2 |
| [Gradle releases](https://gradle.org/releases/), [checksums](https://gradle.org/release-checksums/) | 9.9.2 released Nov 19, 2026; `gradle-9.9.2-bin.zip` SHA-256 `971fcd7150ca4ba04852c0b2ca0d2d0ab9468e14b03473eab19528809914c76b`; wrapper JAR SHA-256 `621008afc4f5d59357fcf06ffa644dd5aade5e091bdc363742710445f73cdaf8`. 9.9.3 (Nov 27) is newer |
| [Gradle wrapper](https://docs.gradle.org/current/userguide/gradle_wrapper.html) | "you need to check them into version control"; "Adding the JAR file to version control is expected."; generating it "requires an installed version of Gradle"; `distributionSha256Sum` makes "a build failure if the configured checksum does not match" |
| [Compose compiler](https://developer.android.com/develop/ui/compose/compiler) | Apply `org.jetbrains.kotlin.plugin.compose` at the Kotlin version; "starting with Compose 1.12.0, projects are required to use `compileSdk 37` and Android Gradle Plugin (AGP) 9." |

## CI

| Source | Text |
|---|---|
| [actions/setup-java](https://github.com/actions/setup-java/releases) | v6.3.0 = `d1374f8ac74474200d16ef8ecee22c7ae7a51880`; `distribution: temurin` |
| [gradle/actions](https://github.com/gradle/actions/releases) | v6.4.1 = `8da46b539cd8fe6ab27ca75128919a18d757604c`. setup-gradle "can download and install a specified Gradle version" (`gradle-version`); "Avoid using `actions/setup-java` with the `cache: gradle` option"; from v4 it will "automatically perform wrapper validation ... on each execution" |
| [actions/checkout](https://github.com/actions/checkout/releases) | v7.0.1 = `3d3c42e5aac5ba805825da76410c181273ba90b1` (already used in this repo) |
| [Now in Android Build.yaml](https://github.com/android/nowinandroid/blob/main/.github/workflows/Build.yaml) | `ubuntu-latest`; setup-java, setup-gradle; runs `:build-logic:convention:check`, unit tests, `:app:assemble`, lint |

## Settled by the sources

1. Convention plugins configure `ApplicationExtension` and `LibraryExtension` and never apply `org.jetbrains.kotlin.android`.
2. Gradle 9.9.2 satisfies AGP 9.5.1 (it needs 9.6.0 or later), so the pin stands.
3. The wrapper (scripts, properties with `distributionSha256Sum`, JAR) belongs in git. Producing it needs a Gradle install, which this workstation lacks, so the first PR runs Gradle 9.9.2 through `setup-gradle` and adds the wrapper later.
