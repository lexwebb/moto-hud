plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "com.motohud.companion"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.motohud.companion"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "0.1.0"
        multiDexEnabled = true
        // Play Feature Delivery requires the on-demand module title in the *base* resource table.
        resValue("string", "osmand_module_title", "OsmAnd rich navigation")
    }

    buildTypes {
        release {
            isMinifyEnabled = false
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
        isCoreLibraryDesugaringEnabled = true
    }
    kotlinOptions {
        jvmTarget = "17"
    }
    buildFeatures {
        viewBinding = true
        buildConfig = true
    }

    // On-demand OsmAnd Full Library (lanes / then-next). Base APK stays small.
    dynamicFeatures += setOf(":osmand")

    bundle {
        abi {
            enableSplit = true
        }
        language {
            enableSplit = true
        }
        density {
            enableSplit = true
        }
    }

    packaging {
        resources {
            pickFirsts += listOf(
                "lib/armeabi-v7a/libc++_shared.so",
                "lib/arm64-v8a/libc++_shared.so",
                "lib/x86_64/libc++_shared.so",
                "lib/x86/libc++_shared.so",
            )
            // Avoid AAB clashes with the :osmand feature module.
            excludes += listOf(
                "META-INF/androidx*.version",
                "META-INF/*.version",
            )
        }
    }
}

dependencies {
    implementation("androidx.core:core-ktx:1.15.0")
    implementation("androidx.appcompat:appcompat:1.7.0")
    implementation("com.google.android.material:material:1.12.0")
    implementation("androidx.constraintlayout:constraintlayout:2.2.0")
    implementation("androidx.multidex:multidex:2.0.1")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.9.0")

    // Play Feature Delivery — download `:osmand` on demand.
    implementation("com.google.android.play:feature-delivery:2.1.0")
    implementation("com.google.android.play:feature-delivery-ktx:2.1.0")

    // Typed OsmAnd AIDL (external OsmAnd app) — always in base.
    implementation("net.osmand:android-aidl-lib:5.3@aar")

    // :osmand pulls Picasso; its ContentProvider is merged into the base
    // manifest on sideload, so the class must live in the base APK too.
    implementation("com.squareup.picasso:picasso:2.71828")

    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.3")

    testImplementation("junit:junit:4.13.2")
    testImplementation("org.json:json:20240303")
}
