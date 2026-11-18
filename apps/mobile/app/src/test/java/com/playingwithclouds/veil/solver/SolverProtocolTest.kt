package com.playingwithclouds.veil.solver

import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class SolverProtocolTest {

    @Test
    fun readsTheSdksRequest() {
        val body = """{"cmd":"request.get","url":"https://missav.ws/en","maxTimeout":60000}"""
        val raw = "POST /v1 HTTP/1.1\r\nHost: 127.0.0.1\r\nContent-Type: application/json\r\nContent-Length: ${body.length}\r\n\r\n$body"
        val request = SolverProtocol.parseRequest(SolverProtocol.readBody(raw.byteInputStream()))
        assertEquals(SolveRequest("https://missav.ws/en", 60_000), request)
    }

    @Test
    fun refusesOtherCommands() {
        assertThrows(IllegalArgumentException::class.java) { SolverProtocol.parseRequest("""{"cmd":"sessions.create"}""") }
        assertThrows(IllegalArgumentException::class.java) { SolverProtocol.parseRequest("""{"cmd":"request.get","url":"file:///etc"}""") }
    }

    @Test
    fun answersAsFlareSolverr() {
        val answer = JSONObject(SolverProtocol.solved("https://missav.ws/en", "<html>ok</html>", "UA"))
        assertEquals("ok", answer.getString("status"))
        assertEquals("<html>ok</html>", answer.getJSONObject("solution").getString("response"))
        assertEquals("error", JSONObject(SolverProtocol.failed("nope")).getString("status"))
    }

    @Test
    fun decodesJavascriptStrings() {
        assertEquals("<p class=\"a\">ü</p>", SolverProtocol.decodeJavascriptString("\"<p class=\\\"a\\\">\\u00fc</p>\""))
        assertEquals("", SolverProtocol.decodeJavascriptString("null"))
    }
}
