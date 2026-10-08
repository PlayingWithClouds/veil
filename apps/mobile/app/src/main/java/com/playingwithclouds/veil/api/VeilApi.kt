package com.playingwithclouds.veil.api

import com.apollographql.apollo.ApolloClient
import com.apollographql.apollo.api.ApolloResponse
import com.apollographql.apollo.api.Operation
import com.apollographql.apollo.api.Optional
import com.apollographql.apollo.network.websocket.GraphQLWsProtocol
import com.apollographql.apollo.network.websocket.WebSocketNetworkTransport
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.IOException
import java.util.concurrent.TimeUnit

/** The backend's GraphQL API: queries over HTTP, subscriptions over graphql-ws. */
object VeilApi {

    private const val HEALTH_POLL_MILLISECONDS = 250L
    private const val STARTUP_TIMEOUT_MILLISECONDS = 30_000L
    private const val HEALTH_TIMEOUT_SECONDS = 2L

    private var clientBaseUrl: String? = null
    private var cachedClient: ApolloClient? = null

    /** Short timeouts, so a dead server address fails fast. */
    private val healthHttpClient = OkHttpClient.Builder()
        .connectTimeout(HEALTH_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .readTimeout(HEALTH_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    /** The client for the server address in use; rebuilt when Settings changes it. */
    val client: ApolloClient
        @Synchronized get() {
            val baseUrl = ServerSettings.baseUrl
            val existing = cachedClient
            if (existing != null && clientBaseUrl == baseUrl) {
                return existing
            }
            existing?.close()
            val created = createClient(baseUrl)
            cachedClient = created
            clientBaseUrl = baseUrl
            return created
        }

    /** Builds a client that talks to the backend at the given address. */
    private fun createClient(baseUrl: String): ApolloClient {
        return ApolloClient.Builder()
            .serverUrl("$baseUrl/graphql")
            .subscriptionNetworkTransport(
                WebSocketNetworkTransport.Builder()
                    .serverUrl(BackendUrls.webSocketUrl(baseUrl))
                    .wsProtocol(GraphQLWsProtocol())
                    .build(),
            )
            .build()
    }

    /**
     * Suspends until the backend at the address in use answers /health, which takes a moment
     * after the app starts. Returns false if it does not come up in time.
     */
    suspend fun awaitBackend(timeoutMilliseconds: Long = STARTUP_TIMEOUT_MILLISECONDS): Boolean {
        val healthy = withTimeoutOrNull(timeoutMilliseconds) {
            while (!isHealthy(ServerSettings.baseUrl)) {
                delay(HEALTH_POLL_MILLISECONDS)
            }
            true
        }
        return healthy == true
    }

    /** Whether a backend answers its health check at the given address. */
    suspend fun isHealthy(url: String): Boolean = withContext(Dispatchers.IO) {
        val request = try {
            Request.Builder().url("${BackendUrls.normalizeServerUrl(url)}/health").build()
        } catch (error: IllegalArgumentException) {
            return@withContext false
        }
        try {
            healthHttpClient.newCall(request).execute().use { response -> response.isSuccessful }
        } catch (error: IOException) {
            false
        }
    }
}

/** The response's data, or an exception carrying the server's error message. */
fun <D : Operation.Data> ApolloResponse<D>.dataOrThrow(): D {
    val graphqlError = errors?.firstOrNull()
    if (graphqlError != null) {
        throw IllegalStateException(graphqlError.message)
    }
    val responseData = data
    if (responseData != null) {
        return responseData
    }
    val failure = exception
    if (failure != null) {
        throw failure
    }
    throw IllegalStateException("empty response")
}

/** An optional GraphQL variable that is left out of the request when the value is null. */
fun <T : Any> T?.orAbsent(): Optional<T?> {
    return Optional.presentIfNotNull(this)
}

/** A GraphQL list variable that is left out of the request when the list is empty. */
fun List<String>.orAbsentIfEmpty(): Optional<List<String>?> {
    if (isEmpty()) {
        return Optional.Absent
    }
    return Optional.present(this)
}
