package com.playingwithclouds.veil.solver

import android.annotation.SuppressLint
import android.content.Context
import android.content.Intent
import android.content.MutableContextWrapper
import android.os.Handler
import android.os.Looper
import android.webkit.CookieManager
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.ProcessLifecycleOwner
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.withLock

/** How often a loaded page is checked for whether Cloudflare's challenge is still up. */
private const val CHECK_INTERVAL_MILLISECONDS = 750L

/** How long the challenge may run by itself before the user is asked to tap it (Turnstile). */
private const val INTERACTIVE_AFTER_MILLISECONDS = 12_000L

/** True while the page is still Cloudflare's interstitial rather than the site. */
private const val CHALLENGE_CHECK = """(function(){
  var title = document.title || '';
  if (title.indexOf('Just a moment') >= 0 || title.indexOf('Attention Required') >= 0) return true;
  return !!document.querySelector('#challenge-form, #challenge-stage, #cf-challenge-running, .cf-turnstile');
})()"""

/** The rendered page, as FlareSolverr returns it. */
private const val PAGE_HTML = "document.documentElement ? document.documentElement.outerHTML : ''"

/**
 * Reads pages behind Cloudflare's browser check the way FlareSolverr does, with the phone's own
 * browser engine: a hidden WebView loads the page, lets the challenge script run and returns the
 * rendered HTML. The clearance cookie stays in the WebView's cookie store, so later pages of the
 * same site load straight through until it expires or the network (IP) changes. When the check
 * wants a tap ("Verify you are human") and the app is in front, [ChallengeActivity] shows the page.
 * One page at a time; the WebView lives on the main thread.
 */
object WebSolver {

    private val lock = ReentrantLock()
    private val mainHandler = Handler(Looper.getMainLooper())
    private lateinit var appContext: Context
    private var contextWrapper: MutableContextWrapper? = null
    private var webView: WebView? = null
    private var userAgent = ""

    private val mutableShown = MutableStateFlow<WebView?>(null)

    /** The WebView while a check needs the user's tap; [ChallengeActivity] shows it. */
    val shown: StateFlow<WebView?> = mutableShown

    /** Remembers the application; called once from the application. */
    fun initialize(context: Context) {
        appContext = context.applicationContext
    }

    /** Loads [request]'s page and returns its HTML once Cloudflare lets it through; throws on timeout. */
    fun fetch(request: SolveRequest): Pair<String, String> {
        lock.withLock {
            val result = CompletableFuture<String>()
            val startedAt = System.currentTimeMillis()
            mainHandler.post { load(request.url, result, startedAt, request.timeoutMilliseconds) }
            try {
                val html = result.get(request.timeoutMilliseconds, TimeUnit.MILLISECONDS)
                return html to userAgent
            } finally {
                mainHandler.post { hide() }
            }
        }
    }

    /** Starts loading [url] and keeps checking until the challenge is gone or time runs out. */
    private fun load(url: String, result: CompletableFuture<String>, startedAt: Long, timeoutMilliseconds: Long) {
        val view = webViewOrCreate()
        view.webViewClient = object : WebViewClient() {
            override fun onPageFinished(view: WebView, finishedUrl: String) {
                check(view, result, startedAt, timeoutMilliseconds)
            }
        }
        view.loadUrl(url)
    }

    /** Reads the page when the challenge is gone; otherwise checks again, asking for a tap after a while. */
    private fun check(view: WebView, result: CompletableFuture<String>, startedAt: Long, timeoutMilliseconds: Long) {
        if (result.isDone) {
            return
        }
        view.evaluateJavascript(CHALLENGE_CHECK) { challenged ->
            if (challenged != "true") {
                view.evaluateJavascript(PAGE_HTML) { html -> result.complete(SolverProtocol.decodeJavascriptString(html)) }
                return@evaluateJavascript
            }
            val elapsed = System.currentTimeMillis() - startedAt
            if (elapsed >= timeoutMilliseconds) {
                result.completeExceptionally(IllegalStateException("Cloudflare's check didn't clear in time"))
                return@evaluateJavascript
            }
            if (elapsed >= INTERACTIVE_AFTER_MILLISECONDS) {
                askForTap(view)
            }
            mainHandler.postDelayed({ check(view, result, startedAt, timeoutMilliseconds) }, CHECK_INTERVAL_MILLISECONDS)
        }
    }

    /** Shows the page so the user can tap the check, when the app is in front and it isn't shown yet. */
    private fun askForTap(view: WebView) {
        if (mutableShown.value != null) {
            return
        }
        if (!ProcessLifecycleOwner.get().lifecycle.currentState.isAtLeast(Lifecycle.State.STARTED)) {
            return
        }
        mutableShown.value = view
        appContext.startActivity(Intent(appContext, ChallengeActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK))
    }

    /** Takes the page off the screen again. */
    private fun hide() {
        mutableShown.value = null
    }

    /** Lets an activity host the WebView (its context is swapped in), or gives it back to the application. */
    fun attachTo(context: Context?) {
        contextWrapper?.baseContext = context ?: appContext
    }

    /** The one WebView, made on first use with JavaScript and a normal mobile browser identity. */
    @SuppressLint("SetJavaScriptEnabled")
    private fun webViewOrCreate(): WebView {
        val existing = webView
        if (existing != null) {
            return existing
        }
        val wrapper = MutableContextWrapper(appContext)
        val created = WebView(wrapper)
        created.settings.javaScriptEnabled = true
        created.settings.domStorageEnabled = true
        created.settings.cacheMode = WebSettings.LOAD_DEFAULT
        CookieManager.getInstance().setAcceptCookie(true)
        CookieManager.getInstance().setAcceptThirdPartyCookies(created, true)
        userAgent = created.settings.userAgentString
        contextWrapper = wrapper
        webView = created
        return created
    }
}
