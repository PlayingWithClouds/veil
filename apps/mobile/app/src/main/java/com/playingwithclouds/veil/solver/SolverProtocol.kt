package com.playingwithclouds.veil.solver

import org.json.JSONArray
import org.json.JSONObject
import java.io.BufferedInputStream
import java.io.ByteArrayOutputStream
import java.io.IOException
import java.io.InputStream

/** The longest a solve may take when the request doesn't say. */
const val DEFAULT_SOLVE_TIMEOUT_MILLISECONDS = 60_000L

/** A page the backend wants read through the browser: FlareSolverr's `request.get`. */
data class SolveRequest(val url: String, val timeoutMilliseconds: Long)

/**
 * FlareSolverr's HTTP API, the part the plugin SDK uses (`fetchViaSolver` in
 * packages/veil-sdk/src/http.ts): a POST of `{cmd: "request.get", url, maxTimeout}` answered with
 * `{status: "ok", solution: {url, status, response}}` or `{status: "error", message}`. Kept free of
 * Android types so it can be tested.
 */
object SolverProtocol {

    private const val MAX_HEADER_BYTES = 16 * 1024
    private const val MAX_BODY_BYTES = 64 * 1024

    /** Reads one HTTP request's body from [input]: the header lines, then Content-Length bytes. */
    fun readBody(input: InputStream): String {
        val buffered = BufferedInputStream(input)
        var contentLength = 0
        while (true) {
            val line = readLine(buffered)
            if (line.isEmpty()) {
                break
            }
            val (name, value) = line.split(":", limit = 2).let { parts -> parts[0] to parts.getOrElse(1) { "" } }
            if (name.trim().equals("Content-Length", ignoreCase = true)) {
                contentLength = value.trim().toIntOrNull() ?: 0
            }
        }
        if (contentLength < 0 || contentLength > MAX_BODY_BYTES) {
            throw IOException("Request body too large")
        }
        val body = ByteArray(contentLength)
        var read = 0
        while (read < contentLength) {
            val count = buffered.read(body, read, contentLength - read)
            if (count < 0) {
                throw IOException("Request body cut short")
            }
            read += count
        }
        return body.decodeToString()
    }

    /** One header line without its CRLF; empty at the blank line ending the headers. */
    private fun readLine(input: InputStream): String {
        val line = ByteArrayOutputStream()
        while (line.size() < MAX_HEADER_BYTES) {
            val next = input.read()
            if (next < 0 || next == '\n'.code) {
                break
            }
            if (next != '\r'.code) {
                line.write(next)
            }
        }
        return line.toString(Charsets.UTF_8.name())
    }

    /** The request in a body; throws [IllegalArgumentException] for anything but `request.get` with a URL. */
    fun parseRequest(body: String): SolveRequest {
        val json = JSONObject(body)
        val command = json.optString("cmd")
        require(command == "request.get") { "Only request.get is supported, not $command" }
        val url = json.optString("url")
        require(url.startsWith("http://") || url.startsWith("https://")) { "A request.get needs an http(s) url" }
        var timeout = json.optLong("maxTimeout", DEFAULT_SOLVE_TIMEOUT_MILLISECONDS)
        if (timeout <= 0) {
            timeout = DEFAULT_SOLVE_TIMEOUT_MILLISECONDS
        }
        return SolveRequest(url, timeout)
    }

    /** The answer for a solved page. */
    fun solved(url: String, html: String, userAgent: String): String {
        val solution = JSONObject()
            .put("url", url)
            .put("status", 200)
            .put("response", html)
            .put("userAgent", userAgent)
            .put("cookies", JSONArray())
        return JSONObject().put("status", "ok").put("message", "").put("solution", solution).toString()
    }

    /** The answer for a page that couldn't be read. */
    fun failed(message: String): String {
        return JSONObject().put("status", "error").put("message", message).toString()
    }

    /** A whole HTTP response with a JSON body. */
    fun httpResponse(statusCode: Int, json: String): ByteArray {
        val body = json.toByteArray()
        var reason = "OK"
        if (statusCode != 200) {
            reason = "Error"
        }
        val head = "HTTP/1.1 $statusCode $reason\r\nContent-Type: application/json\r\nContent-Length: ${body.size}\r\nConnection: close\r\n\r\n"
        return head.toByteArray() + body
    }

    /** A string evaluateJavascript returned (JSON-encoded) as plain text; "" for null or undefined. */
    fun decodeJavascriptString(value: String?): String {
        if (value == null || value == "null" || value == "undefined") {
            return ""
        }
        return JSONArray("[$value]").optString(0, "")
    }
}
