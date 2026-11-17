package com.playingwithclouds.veil.ui.home

import android.content.Context
import android.graphics.Bitmap
import android.media.FaceDetector
import coil3.SingletonImageLoader
import coil3.request.ImageRequest
import coil3.request.SuccessResult
import coil3.request.allowHardware
import coil3.size.Size
import coil3.toBitmap
import com.playingwithclouds.veil.api.ServerSettings
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

/** Faces looked for in a cover; more than this many people makes a busier picture than wanted. */
private const val MAX_FACES = 3

/** Width covers are read at: plenty for finding a face, cheap to scan. */
private const val SCAN_WIDTH = 480

/**
 * Whether a cover shows one or two faces, found by Android's built-in detector (it finds faces by
 * their eyes). Faces with eye contact and one or two people hold attention best (Netflix artwork
 * tests, Rupp & Wallen's eye tracking), so Home leads with such covers. False when the cover can't
 * be read.
 */
suspend fun coverShowsFaces(context: Context, coverUrl: String?): Boolean = withContext(Dispatchers.Default) {
    val url = ServerSettings.imageUrl(coverUrl) ?: return@withContext false
    val request = ImageRequest.Builder(context).data(url).allowHardware(false).size(Size(SCAN_WIDTH, SCAN_WIDTH)).build()
    val result = SingletonImageLoader.get(context).execute(request) as? SuccessResult ?: return@withContext false
    val faces = runCatching { countFaces(result.image.toBitmap()) }.getOrDefault(0)
    faces in 1..2
}

/** Faces the detector finds in a bitmap, up to [MAX_FACES]. */
private fun countFaces(bitmap: Bitmap): Int {
    // The detector wants RGB 565 and an even width.
    val evenWidth = bitmap.width - bitmap.width % 2
    val scan = Bitmap.createBitmap(bitmap, 0, 0, evenWidth, bitmap.height).copy(Bitmap.Config.RGB_565, false)
    val found = arrayOfNulls<FaceDetector.Face>(MAX_FACES)
    return FaceDetector(scan.width, scan.height, MAX_FACES).findFaces(scan, found)
}
