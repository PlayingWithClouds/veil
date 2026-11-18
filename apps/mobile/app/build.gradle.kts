plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.apollo)
}

// Version from `-PveilVersion=1.2.3` (CI passes the release tag); local builds keep the default.
val defaultVersionName = "0.0.1"

/** Reads the version name from the Gradle property, or the default for local builds. */
fun veilVersionName(): String {
    val property = project.findProperty("veilVersion") as String?
    if (property == null) {
        return defaultVersionName
    }
    return property.removePrefix("v")
}

/** Derives a monotonic version code from `major.minor.patch`: 1.2.3 → 1_002_003. */
fun veilVersionCode(versionName: String): Int {
    val parts = versionName.substringBefore("-").split(".").map { it.toInt() }
    require(parts.size == 3) { "veilVersion must be major.minor.patch, got $versionName" }
    return parts[0] * 1_000_000 + parts[1] * 1_000 + parts[2]
}

/**
 * The scene model export to bundle: `-PsceneModelDir=…` or `VEIL_SCENE_MODEL_DIR`, else the newest
 * `runs/<export>/onnx` of the classifier repo beside this one (by its manifest's date); null when
 * none is there, so other machines and CI build without a model.
 */
fun sceneModelExport(): File? {
    val configured = (project.findProperty("sceneModelDir") as String?) ?: System.getenv("VEIL_SCENE_MODEL_DIR")
    if (configured != null) {
        return file(configured).takeIf { directory -> File(directory, "manifest.json").isFile }
    }
    val runs = rootDir.resolve("../../../nsfw-scene-recognition/runs")
    val manifests = runs.listFiles().orEmpty().map { run -> File(run, "onnx/manifest.json") }.filter { manifest -> manifest.isFile }
    return manifests.maxByOrNull { manifest -> manifest.lastModified() }?.parentFile
}

/** The files the export's manifest.json names (the frame and temporal graphs), with the manifest itself. */
fun sceneModelFiles(export: File?): List<File> {
    if (export == null) {
        return emptyList()
    }
    val manifest = File(export, "manifest.json")
    val graphNames = Regex("\"(frame_model|temporal_model)\"\\s*:\\s*\"([^\"]+)\"").findAll(manifest.readText()).map { match -> match.groupValues[2] }
    return listOf(manifest) + graphNames.map { name -> File(export, name) }.filter { graph -> graph.isFile }
}

/** Copies the scene model export into generated assets (`scene-model/`); with no export the folder stays empty. */
abstract class BundleSceneModel : DefaultTask() {
    @get:InputFiles
    @get:PathSensitive(PathSensitivity.NAME_ONLY)
    abstract val modelFiles: ConfigurableFileCollection

    @get:OutputDirectory
    abstract val outputDirectory: DirectoryProperty

    /** Replaces the bundled copy with the export's files. */
    @TaskAction
    fun copy() {
        val target = outputDirectory.get().asFile.resolve("scene-model")
        target.deleteRecursively()
        target.mkdirs()
        for (source in modelFiles.files) {
            source.copyTo(File(target, source.name), overwrite = true)
        }
    }
}

val bundleSceneModel = tasks.register<BundleSceneModel>("bundleSceneModel") {
    modelFiles.from(sceneModelFiles(sceneModelExport()))
}

androidComponents {
    onVariants { variant ->
        variant.sources.assets?.addGeneratedSourceDirectory(bundleSceneModel, BundleSceneModel::outputDirectory)
    }
}

// Release signing from the environment (CI secrets). Absent locally, so release builds
// fall back to the debug key there, as before.
val releaseKeystorePath: String? = System.getenv("VEIL_KEYSTORE_PATH")

android {
    namespace = "com.playingwithclouds.veil"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.playingwithclouds.veil"
        minSdk = 26
        targetSdk = 36
        versionName = veilVersionName()
        versionCode = veilVersionCode(versionName!!)
        // Loopback port of the embedded Go backend; the web client mirrors it in lib/server.ts.
        buildConfigField("int", "BACKEND_PORT", "47831")
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    signingConfigs {
        if (releaseKeystorePath != null) {
            create("release") {
                storeFile = file(releaseKeystorePath)
                storePassword = System.getenv("VEIL_KEYSTORE_PASSWORD")
                keyAlias = System.getenv("VEIL_KEY_ALIAS")
                keyPassword = System.getenv("VEIL_KEY_PASSWORD")
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            // Published builds use the release key; personal local builds the debug key,
            // which lets them install over debug builds.
            if (releaseKeystorePath != null) {
                signingConfig = signingConfigs.getByName("release")
            } else {
                signingConfig = signingConfigs.getByName("debug")
            }
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
    implementation(libs.compose.ui.tooling.preview)
    debugImplementation(libs.compose.ui.tooling)
    implementation(libs.activity.compose)
    implementation(libs.navigation.compose)
    implementation(libs.lifecycle.runtime.compose)
    implementation(libs.lifecycle.viewmodel.compose)
    implementation(libs.lifecycle.process)
    implementation(libs.work.runtime)
    implementation(libs.media3.exoplayer)
    implementation(libs.media3.exoplayer.hls)
    implementation(libs.media3.ui)
    implementation(libs.media3.ui.compose)
    implementation(libs.media3.inspector.frame)
    implementation(libs.onnxruntime.android)
    implementation(libs.apollo.runtime)
    implementation(libs.okhttp)
    implementation(libs.coil.compose)
    implementation(libs.coil.network)
    implementation(libs.haze)
    implementation(libs.haze.blur)
    implementation(libs.phosphor)
    implementation(libs.biometric)

    testImplementation(libs.junit)
    testImplementation(libs.coroutines.test)
    testImplementation(libs.json)
}
