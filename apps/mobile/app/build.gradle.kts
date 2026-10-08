plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.apollo)
}

android {
    namespace = "com.playingwithclouds.veil"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.playingwithclouds.veil"
        minSdk = 26
        targetSdk = 36
        versionCode = 2
        versionName = "2.0.0"
        // Loopback port of the embedded Go backend; the web client mirrors it in lib/server.ts.
        buildConfigField("int", "BACKEND_PORT", "47831")
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            // Sideloaded personal builds: the debug key lets them install over debug builds.
            signingConfig = signingConfigs.getByName("debug")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    // libveil.so is the Go backend executable, not a library: it must be
    // extracted to nativeLibraryDir to be runnable.
    packaging {
        jniLibs {
            useLegacyPackaging = true
        }
    }

    testOptions {
        unitTests.isReturnDefaultValues = true
    }
}

apollo {
    service("veil") {
        packageName.set("com.playingwithclouds.veil.graphql")
        // The backend's schema is split per entity (`extend type Query`); Apollo merges the files.
        schemaFiles.from(fileTree("../../backend/internal/api/graphql/schema") { include("*.graphql") })
    }
}

dependencies {
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.material3)
    implementation(libs.compose.material.icons.core)
    implementation(libs.compose.material.icons.extended)
    implementation(libs.compose.ui.tooling.preview)
    debugImplementation(libs.compose.ui.tooling)
    implementation(libs.activity.compose)
    implementation(libs.navigation.compose)
    implementation(libs.lifecycle.runtime.compose)
    implementation(libs.lifecycle.viewmodel.compose)
    implementation(libs.media3.exoplayer)
    implementation(libs.media3.exoplayer.hls)
    implementation(libs.media3.ui)
    implementation(libs.media3.ui.compose)
    implementation(libs.apollo.runtime)
    implementation(libs.okhttp)
    implementation(libs.coil.compose)
    implementation(libs.coil.network)

    testImplementation(libs.junit)
    testImplementation(libs.coroutines.test)
    testImplementation(libs.json)
}
