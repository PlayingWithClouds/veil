package com.playingwithclouds.veil.backend

/**
 * How the backend protects downloads and cached streams at rest: the base64 AES-256 key (always
 * given, so earlier encrypted blobs stay readable) and whether new blobs are encrypted.
 */
data class BlobEncryption(val encodedKey: String, val encryptNewBlobs: Boolean)

/** What the Go backend is started with and which files it reads, kept free of Android types so it can be tested. */
object BackendEnvironment {

    /** Used when the network reports no DNS servers of its own. */
    const val FALLBACK_DNS_SERVERS = "1.1.1.1,8.8.8.8"

    /** The address the embedded backend listens on; it binds loopback only. */
    fun loopbackUrl(port: Int): String {
        return "http://127.0.0.1:$port"
    }

    /** Environment variables for the backend process. */
    fun variables(
        dataDir: String,
        pluginSeedDir: String,
        dnsFile: String,
        tempDir: String,
        port: Int,
        blobEncryption: BlobEncryption? = null,
    ): Map<String, String> {
        val variables = mutableMapOf(
            "DATA_DIR" to dataDir,
            "HOST" to "127.0.0.1",
            "PORT" to port.toString(),
            "PUBLIC_URL" to loopbackUrl(port),
            "PLUGIN_SEED_DIR" to pluginSeedDir,
            "DNS_SERVERS_FILE" to dnsFile,
            "DNS_SERVERS" to FALLBACK_DNS_SERVERS,
            // Go's os.TempDir defaults to /data/local/tmp, which apps can't write.
            "TMPDIR" to tempDir,
        )
        if (blobEncryption != null) {
            variables["BLOB_ENCRYPTION_KEY"] = blobEncryption.encodedKey
            if (!blobEncryption.encryptNewBlobs) {
                variables["BLOB_ENCRYPTION_WRITES"] = "off"
            }
        }
        return variables
    }

    /** The DNS server file's text: one server per line, as the backend re-reads it on every lookup. */
    fun dnsServerFileContents(servers: List<String>): String {
        return servers.joinToString("\n")
    }

    /** How long to wait before restarting a crashed backend: longer after every crash. */
    fun restartDelayMilliseconds(restartCount: Int): Long {
        return 1000L * restartCount
    }
}
