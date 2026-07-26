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
        // OsmAnd AIDL stubs (android-aidl-lib) — same Ivy repo as osmand-api-demo
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
    }
}
rootProject.name = "MotoHUD"
include(":app")
