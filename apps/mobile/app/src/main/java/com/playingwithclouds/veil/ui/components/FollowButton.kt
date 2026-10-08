package com.playingwithclouds.veil.ui.components

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.data.FollowKind
import com.playingwithclouds.veil.data.SubscriptionSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.launch

/** Follows a studio, performer or tag; once followed, opens the subscription. */
@Composable
fun FollowButton(kind: FollowKind, targetId: String, navigator: AppNavigator, modifier: Modifier = Modifier) {
    var subscription by remember(targetId) { mutableStateOf<SubscriptionSummary?>(null) }
    var busy by remember(targetId) { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    LaunchedEffect(targetId) {
        try {
            subscription = EntityRepository.subscriptionFor(targetId)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            subscription = null
        }
    }

    val followed = subscription
    if (followed != null) {
        SecondaryButton(
            "Following",
            onClick = { navigator.openSubscription(followed.id) },
            modifier = modifier,
            icon = VeilIcons.Check,
        )
        return
    }
    PrimaryButton(
        "Follow",
        enabled = !busy,
        icon = VeilIcons.Follow,
        onClick = {
            scope.launch {
                busy = true
                try {
                    subscription = EntityRepository.follow(kind, targetId)
                } catch (error: CancellationException) {
                    throw error
                } catch (error: Exception) {
                    subscription = null
                }
                busy = false
            }
        },
        modifier = modifier,
    )
}
