package com.playingwithclouds.veil.backend

import org.junit.Assert.assertEquals
import org.junit.Test

class BackendEnvironmentTest {

    @Test
    fun variablesPointTheBackendAtItsFoldersAndLoopback() {
        val variables = BackendEnvironment.variables(
            dataDir = "/data/files/veil",
            pluginSeedDir = "/data/files/plugin-seed",
            dnsFile = "/data/files/veil/dns-servers",
            tempDir = "/data/cache",
            port = 47831,
        )
        assertEquals("/data/files/veil", variables["DATA_DIR"])
        assertEquals("127.0.0.1", variables["HOST"])
        assertEquals("47831", variables["PORT"])
        assertEquals("http://127.0.0.1:47831", variables["PUBLIC_URL"])
        assertEquals("/data/files/plugin-seed", variables["PLUGIN_SEED_DIR"])
        assertEquals("/data/files/veil/dns-servers", variables["DNS_SERVERS_FILE"])
        assertEquals("1.1.1.1,8.8.8.8", variables["DNS_SERVERS"])
        assertEquals("/data/cache", variables["TMPDIR"])
    }

    @Test
    fun dnsFileHasOneServerPerLine() {
        assertEquals("192.168.1.1\nfe80::1", BackendEnvironment.dnsServerFileContents(listOf("192.168.1.1", "fe80::1")))
    }

    @Test
    fun dnsFileOfNoServersIsEmpty() {
        assertEquals("", BackendEnvironment.dnsServerFileContents(emptyList()))
    }

    @Test
    fun restartDelayGrowsWithEveryCrash() {
        assertEquals(1000L, BackendEnvironment.restartDelayMilliseconds(1))
        assertEquals(3000L, BackendEnvironment.restartDelayMilliseconds(3))
    }

    @Test
    fun loopbackUrlUsesThePort() {
        assertEquals("http://127.0.0.1:1234", BackendEnvironment.loopbackUrl(1234))
    }
}
