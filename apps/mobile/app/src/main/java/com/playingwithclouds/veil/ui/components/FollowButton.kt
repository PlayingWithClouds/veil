package com.playingwithclouds.veil.ui.components

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.NotificationsActive
import androidx.compose.material3.Button
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.layout.size
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.data.FollowKind
import com.playingwithclouds.veil.data.SubscriptionSummary
import com.playingwithclouds.veil.ui.AppNavigator
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
        OutlinedButton(onClick = { navigator.openSubscription(followed.id) }, modifier = modifier) {
            Icon(Icons.Filled.Check, contentDescription = null, modifier = Modifier.size(18.dp))
            Text("Following")
        }
        return
    }
    Button(
        enabled = !busy,
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
    ) {
        Icon(Icons.Filled.NotificationsActive, contentDescription = null, modifier = Modifier.size(18.dp))
        Text("Follow")
    }
}
