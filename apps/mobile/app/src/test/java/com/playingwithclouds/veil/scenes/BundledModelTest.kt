package com.playingwithclouds.veil.scenes

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class BundledModelTest {

    @Test
    fun installsWhenNothingIsInstalled() {
        assertTrue(BundledModel.isDue(null, "export-a", null))
        assertTrue(BundledModel.isDue(null, "export-a", "export-a"))
    }

    @Test
    fun skipsWhatIsAlreadyInstalled() {
        assertFalse(BundledModel.isDue("export-a", "export-a", null))
    }

    @Test
    fun aUserModelStaysUntilAnUpdateBringsAnotherExport() {
        assertFalse(BundledModel.isDue("from-url", "export-a", "export-a"))
        assertTrue(BundledModel.isDue("from-url", "export-b", "export-a"))
    }

    @Test
    fun anUpdateReplacesTheOlderBundledModel() {
        assertTrue(BundledModel.isDue("export-a", "export-b", "export-a"))
    }
}
