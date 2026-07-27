plugins {
    id("com.android.dynamic-feature")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "com.motohud.companion.osmand"
    compileSdk = 35

    defaultConfig {
        minSdk = 26
        // Phones in the wild are arm64; emulators needing x86 use the lite AIDL path.
        ndk {
            abiFilters += listOf("arm64-v8a")
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
        buildConfig = true
    }
    packaging {
        resources {
            pickFirsts += listOf(
                "lib/armeabi-v7a/libc++_shared.so",
                "lib/arm64-v8a/libc++_shared.so",
                "lib/x86_64/libc++_shared.so",
                "lib/x86/libc++_shared.so",
                "META-INF/androidx*.version",
                "META-INF/*.kotlin_module",
            )
            // Maps download at runtime; shipping the mini basemap costs ~30–60 MB forever.
            // Also drop duplicate AndroidX version stamps that clash with the base module in AABs.
            excludes += listOf(
                "assets/World_basemap_mini.obf",
                "**/World_basemap_mini.obf",
                "META-INF/androidx*.version",
                "META-INF/*.version",
            )
        }
        jniLibs {
            excludes += listOf(
                "**/armeabi-v7a/**",
                "**/x86/**",
                "**/x86_64/**",
            )
        }
    }
    androidResources {
        noCompress += listOf("qz", "png")
        // Default ignore pattern + drop the bundled mini basemap (~64 MB); regions download at runtime.
        ignoreAssetsPattern =
            "!.svn:!.git:!.ds_store:!*.scc:.*:!CVS:!thumbs.db:!picasa.ini:!*~:World_basemap_mini.obf"
    }
}

// Belt-and-braces: delete mini basemap from merged assets if AAPT still pulled it in.
tasks.configureEach {
    if (name.startsWith("merge") && name.contains("Assets", ignoreCase = true)) {
        doLast {
            val out = outputs.files.files
            out.forEach { dir ->
                fileTree(dir).matching { include("**/World_basemap_mini.obf") }.forEach { f ->
                    logger.lifecycle("Stripping ${f.name} from $path")
                    f.delete()
                }
            }
        }
    }
}

dependencies {
    implementation(project(":app"))
    implementation("com.google.android.play:feature-delivery:2.1.0")

    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.3")

    // OsmAnd Full Library — mirrors osmand-api-demo OsmAnd-map-sample (trimmed).
    implementation("net.osmand:android-aidl-lib:5.3@aar")
    implementation("net.osmand:OsmAnd-java:master-snapshot:android@jar")
    add("debugImplementation", "net.osmand:OsmAnd:master-snapshot:debug@aar")
    add("releaseImplementation", "net.osmand:OsmAnd:master-snapshot:release@aar")
    add("debugImplementation", "net.osmand.shared:OsmAnd-shared-android:master-snapshot:debug@aar")
    add("releaseImplementation", "net.osmand.shared:OsmAnd-shared-android:master-snapshot:release@aar")
    add("debugImplementation", "net.osmand:MPAndroidChart:custom-snapshot-debug@aar")
    add("releaseImplementation", "net.osmand:MPAndroidChart:custom-snapshot-release@aar")
    implementation("net.osmand:OsmAndCore_androidNativeRelease:master-snapshot@aar")
    implementation("net.osmand:OsmAndCore_android:master-snapshot@aar")

    implementation("androidx.multidex:multidex:2.0.1")
    implementation("androidx.preference:preference-ktx:1.2.1")
    implementation("androidx.lifecycle:lifecycle-process:2.8.7")
    implementation("androidx.cardview:cardview:1.0.0")
    implementation("androidx.gridlayout:gridlayout:1.0.0")
    implementation("androidx.browser:browser:1.8.0")
    implementation("androidx.activity:activity:1.10.1")
    implementation("com.google.android.gms:play-services-location:21.3.0")
    implementation("com.google.android.play:review:2.0.1")
    implementation("com.android.billingclient:billing:8.0.0")

    implementation("commons-logging:commons-logging:1.2")
    implementation("commons-codec:commons-codec:1.11")
    implementation("org.apache.commons:commons-compress:1.17")
    implementation("com.moparisthebest:junidecode:0.1.1")
    implementation("org.immutables:gson:2.5.0")
    implementation("com.vividsolutions:jts-core:1.14.0")
    implementation("com.google.openlocationcode:openlocationcode:1.0.4")
    implementation("org.mozilla:rhino:1.7.9")
    implementation("com.squareup.picasso:picasso:2.71828")
    implementation("me.zhanghai.android.materialprogressbar:library:1.4.2")
    implementation("com.getkeepsafe.taptargetview:taptargetview:1.13.2") {
        exclude(group = "com.android.support")
    }
    implementation("com.github.HITGIF:TextFieldBoxes:1.4.5") {
        exclude(group = "com.android.support")
    }
    implementation("com.github.scribejava:scribejava-apis:7.1.1") {
        exclude(group = "com.fasterxml.jackson.core")
    }
    implementation("com.jaredrummler:colorpicker:1.1.0")
    implementation("net.osmand:antpluginlib:3.8.0@aar")
    implementation("androidx.car.app:app:1.4.0")
    implementation("androidx.car.app:app-projected:1.4.0")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.3")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.1")
    implementation("org.jetbrains.kotlinx:kotlinx-datetime:0.6.0")
    implementation("com.squareup.okio:okio:3.9.0")
    implementation("co.touchlab:stately-concurrent-collections:2.1.0")
    implementation("androidx.sqlite:sqlite:2.3.1")
    implementation("androidx.sqlite:sqlite-framework:2.3.1")
    implementation("net.sf.kxml:kxml2:2.3.0")
    implementation("com.facebook.shimmer:shimmer:0.5.0@aar")
}

configurations.configureEach {
    exclude(group = "com.android.support")
    exclude(module = "support-v4")
    exclude(module = "support-compat")
    exclude(module = "support-media-compat")
}
