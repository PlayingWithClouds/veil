package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.SearchSite
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * "All" and one badge per site with its icon, running edge to edge past a padded grid's gutter;
 * several sites may be picked at once. Hidden when there is nothing to choose between.
 */
@Composable
fun SiteBadges(sites: List<SearchSite>, selected: Set<String>, onToggle: (String) -> Unit, onAll: () -> Unit) {
    if (sites.size < 2) {
        return
    }
    LazyRow(
        Modifier.bleed().fillMaxWidth().fadingEdges().padding(vertical = VeilSpacing.extraSmall),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        item(key = "all") {
            Pill("All", onClick = onAll, selected = selected.isEmpty())
        }
        items(sites, key = { site -> site.name }) { site ->
            Pill(
                site.label,
                onClick = { onToggle(site.name) },
                selected = selected.contains(site.name),
                leading = { SiteIcon(site.iconUrl) },
            )
        }
    }
}

/** A site's small icon inside its badge; nothing when the site has none. */
@Composable
private fun SiteIcon(iconUrl: String?) {
    if (iconUrl == null) {
        return
    }
    RemoteImage(iconUrl, Modifier.size(16.dp).clip(VeilShapes.badge), contentScale = ContentScale.Fit)
}
