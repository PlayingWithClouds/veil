package com.playingwithclouds.veil.solver

import android.util.Log
import java.net.InetAddress
import java.net.ServerSocket
import java.net.Socket
import kotlin.concurrent.thread

private const val TAG = "VeilSolver"

/** How many requests may wait for the browser at once; more are turned away rather than queued forever. */
private const val BACKLOG = 8

/**
 * A FlareSolverr stand-in on loopback for the embedded backend: plugins whose site sits behind
 * Cloudflare's browser check (missav, spankbang) ask it to read a page, and [WebSolver] does it
 * with the phone's browser engine. The backend gets its address as FLARESOLVERR_URL.
 */
object SolverServer {

    private var socket: ServerSocket? = null

    /** The address to give the backend; null until [start] ran. */
    val endpoint: String?
        get() {
            val bound = socket ?: return null
            return "http://127.0.0.1:${bound.localPort}/v1"
        }

    /** Binds a free loopback port and serves requests on a background thread; safe to call twice. */
    fun start() {
        if (socket != null) {
            return
        }
        // IPv4 explicitly: getLoopbackAddress() can be ::1, and the backend dials 127.0.0.1.
        val bound = ServerSocket(0, BACKLOG, InetAddress.getByName("127.0.0.1"))
        socket = bound
        thread(name = "veil-solver", isDaemon = true) {
            while (!bound.isClosed) {
                val client = runCatching { bound.accept() }.getOrNull() ?: continue
                thread(name = "veil-solver-request", isDaemon = true) { serve(client) }
            }
        }
    }

    /** Answers one request: the page through the browser, or why it couldn't be read. */
    private fun serve(client: Socket) {
        client.use { connection ->
            val answer = runCatching {
                val request = SolverProtocol.parseRequest(SolverProtocol.readBody(connection.getInputStream()))
                val (html, userAgent) = WebSolver.fetch(request)
                SolverProtocol.solved(request.url, html, userAgent)
            }
            val body = answer.getOrElse { error ->
                Log.w(TAG, "solve failed: ${error.message}")
                SolverProtocol.failed(error.message ?: "The page couldn't be read")
            }
            connection.getOutputStream().write(SolverProtocol.httpResponse(200, body))
        }
    }
}
