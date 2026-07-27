pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        // OsmAnd AIDL stubs + Full Library AARs — same Ivy repo as osmand-api-demo
        ivy {
            name = "OsmAndBinariesIvy"
            url = uri("https://builder.osmand.net")
            patternLayout {
                artifact("ivy/[organisation]/[module]/[revision]/[artifact]-[revision](-[classifier]).[ext]")
            }
            metadataSources {
                artifact()
            }
        }
        google()
        mavenCentral()
        maven { url = uri("https://jitpack.io") }
    }
}
rootProject.name = "MotoHUD"
include(":app")
include(":osmand")
