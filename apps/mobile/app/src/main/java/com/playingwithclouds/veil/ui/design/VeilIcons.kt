package com.playingwithclouds.veil.ui.design

import androidx.compose.ui.graphics.vector.ImageVector
import com.adamglin.PhosphorIcons
import com.adamglin.phosphoricons.Fill
import com.adamglin.phosphoricons.Regular
import com.adamglin.phosphoricons.fill.BookmarkSimple as BookmarkSimpleFill
import com.adamglin.phosphoricons.fill.Books as BooksFill
import com.adamglin.phosphoricons.fill.Heart as HeartFill
import com.adamglin.phosphoricons.fill.House as HouseFill
import com.adamglin.phosphoricons.fill.Play as PlayFill
import com.adamglin.phosphoricons.fill.Rss as RssFill
import com.adamglin.phosphoricons.fill.ThumbsDown as ThumbsDownFill
import com.adamglin.phosphoricons.fill.ThumbsUp as ThumbsUpFill
import com.adamglin.phosphoricons.regular.ArrowClockwise
import com.adamglin.phosphoricons.regular.ArrowLeft
import com.adamglin.phosphoricons.regular.BellRinging
import com.adamglin.phosphoricons.regular.BookmarkSimple
import com.adamglin.phosphoricons.regular.Books
import com.adamglin.phosphoricons.regular.Buildings
import com.adamglin.phosphoricons.regular.CaretRight
import com.adamglin.phosphoricons.regular.Check
import com.adamglin.phosphoricons.regular.ClockCounterClockwise
import com.adamglin.phosphoricons.regular.DownloadSimple
import com.adamglin.phosphoricons.regular.FolderSimplePlus
import com.adamglin.phosphoricons.regular.GearSix
import com.adamglin.phosphoricons.regular.Heart
import com.adamglin.phosphoricons.regular.House
import com.adamglin.phosphoricons.regular.Images
import com.adamglin.phosphoricons.regular.ListMagnifyingGlass
import com.adamglin.phosphoricons.regular.MagnifyingGlass
import com.adamglin.phosphoricons.regular.PencilSimple
import com.adamglin.phosphoricons.regular.Playlist
import com.adamglin.phosphoricons.regular.Plus
import com.adamglin.phosphoricons.regular.Prohibit
import com.adamglin.phosphoricons.regular.PuzzlePiece
import com.adamglin.phosphoricons.regular.Rss
import com.adamglin.phosphoricons.regular.Shuffle
import com.adamglin.phosphoricons.regular.Tag
import com.adamglin.phosphoricons.regular.ThumbsDown
import com.adamglin.phosphoricons.regular.ThumbsUp
import com.adamglin.phosphoricons.regular.Trash
import com.adamglin.phosphoricons.regular.User
import com.adamglin.phosphoricons.regular.UsersThree
import com.adamglin.phosphoricons.regular.X

/**
 * The app's icon set (Phosphor). Screens take icons from here, never from Material icons, so the
 * look stays consistent; add an entry before using a new one.
 */
object VeilIcons {
    val Alternates: ImageVector get() = PhosphorIcons.Regular.ListMagnifyingGlass
    val Back: ImageVector get() = PhosphorIcons.Regular.ArrowLeft
    val Block: ImageVector get() = PhosphorIcons.Regular.Prohibit
    val Bookmark: ImageVector get() = PhosphorIcons.Regular.BookmarkSimple
    val BookmarkFilled: ImageVector get() = PhosphorIcons.Fill.BookmarkSimpleFill
    val Check: ImageVector get() = PhosphorIcons.Regular.Check
    val Chevron: ImageVector get() = PhosphorIcons.Regular.CaretRight
    val Close: ImageVector get() = PhosphorIcons.Regular.X
    val CollectionAdd: ImageVector get() = PhosphorIcons.Regular.FolderSimplePlus
    val Collections: ImageVector get() = PhosphorIcons.Regular.Playlist
    val Delete: ImageVector get() = PhosphorIcons.Regular.Trash
    val Dislike: ImageVector get() = PhosphorIcons.Regular.ThumbsDown
    val DislikeFilled: ImageVector get() = PhosphorIcons.Fill.ThumbsDownFill
    val Download: ImageVector get() = PhosphorIcons.Regular.DownloadSimple
    val Edit: ImageVector get() = PhosphorIcons.Regular.PencilSimple
    val Follow: ImageVector get() = PhosphorIcons.Regular.BellRinging
    val Following: ImageVector get() = PhosphorIcons.Regular.Rss
    val FollowingFilled: ImageVector get() = PhosphorIcons.Fill.RssFill
    val Galleries: ImageVector get() = PhosphorIcons.Regular.Images
    val Heart: ImageVector get() = PhosphorIcons.Regular.Heart
    val HeartFilled: ImageVector get() = PhosphorIcons.Fill.HeartFill
    val History: ImageVector get() = PhosphorIcons.Regular.ClockCounterClockwise
    val Home: ImageVector get() = PhosphorIcons.Regular.House
    val HomeFilled: ImageVector get() = PhosphorIcons.Fill.HouseFill
    val Library: ImageVector get() = PhosphorIcons.Regular.Books
    val LibraryFilled: ImageVector get() = PhosphorIcons.Fill.BooksFill
    val Like: ImageVector get() = PhosphorIcons.Regular.ThumbsUp
    val LikeFilled: ImageVector get() = PhosphorIcons.Fill.ThumbsUpFill
    val Performer: ImageVector get() = PhosphorIcons.Regular.User
    val Performers: ImageVector get() = PhosphorIcons.Regular.UsersThree
    val Play: ImageVector get() = PhosphorIcons.Fill.PlayFill
    val Plugins: ImageVector get() = PhosphorIcons.Regular.PuzzlePiece
    val Plus: ImageVector get() = PhosphorIcons.Regular.Plus
    val Random: ImageVector get() = PhosphorIcons.Regular.Shuffle
    val Refresh: ImageVector get() = PhosphorIcons.Regular.ArrowClockwise
    val Search: ImageVector get() = PhosphorIcons.Regular.MagnifyingGlass
    val Settings: ImageVector get() = PhosphorIcons.Regular.GearSix
    val Studios: ImageVector get() = PhosphorIcons.Regular.Buildings
    val Tags: ImageVector get() = PhosphorIcons.Regular.Tag
}
