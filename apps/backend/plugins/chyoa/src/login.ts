import { jarFetch, jarOf, loggedIn, newJar } from "@playingwithclouds/veil-sdk";
import type { AccountResult } from "@playingwithclouds/veil-sdk";
import { parse } from "node-html-parser";
import { LOGIN_URL, SITE_URL, isLoginPage, normalizeText, pageHeaders, storyUrl } from "./site.ts";

// A story that sends logged-out visitors to the login page.
const CHECK_STORY_ID = "6006";

const FAILED_MESSAGE = "Chyoa did not accept this username and password";

/** The hidden CSRF token of the login form, or undefined when the page has none. */
export function csrfTokenOf(html: string): string | undefined {
  const token = parse(html).querySelector('input[name="_token"]')?.getAttribute("value");
  if (token === undefined || token === "") return undefined;
  return token;
}

/** The reason the login page gives for turning a login down, else a general message. */
export function loginErrorOf(html: string): string {
  const messages: string[] = [];
  for (const element of parse(html).querySelectorAll(".rd-auth-field-error, .alert, .help-block")) {
    const message = normalizeText(element.text);
    if (message !== "" && !messages.includes(message)) messages.push(message);
  }
  if (messages.length === 0) return FAILED_MESSAGE;
  return messages.join(" ");
}

/** The headers of the login form's POST. */
function postHeaders(): Record<string, string> {
  return {
    ...pageHeaders(),
    "Content-Type": "application/x-www-form-urlencoded",
    Origin: SITE_URL,
    Referer: LOGIN_URL,
  };
}

/**
 * Serves "account:login": reads the login form's CSRF token (and the session
 * cookie that goes with it), posts the credentials and checks that the site
 * took us away from the login page.
 */
export async function login(values: Record<string, string>): Promise<AccountResult> {
  const username = (values.username ?? "").trim();
  const password = values.password ?? "";
  if (username === "" || password === "") {
    return { status: "failed", error: "Enter your username and password" };
  }

  const jar = newJar();
  const formPage = await jarFetch(jar, LOGIN_URL, { headers: pageHeaders() });
  if (!formPage.ok) return { status: "failed", error: `Chyoa's login page answered HTTP ${formPage.status}` };
  const token = csrfTokenOf(await formPage.text());
  if (token === undefined) return { status: "failed", error: "Chyoa's login page has no login form" };

  const body = new URLSearchParams({ _token: token, username, password, remember: "1" });
  const response = await jarFetch(jar, LOGIN_URL, { method: "POST", headers: postHeaders(), body: body.toString() });
  const html = await response.text();
  if (isLoginPage(html)) return { status: "failed", error: loginErrorOf(html) };
  if (!response.ok) return { status: "failed", error: `Chyoa answered HTTP ${response.status}` };
  return loggedIn(jar, username);
}

/**
 * Serves "account:check": whether the session still gets past the login
 * redirect that a story page answers logged-out visitors with.
 */
export async function checkAccount(session: string): Promise<AccountResult> {
  const response = await jarFetch(jarOf(session), storyUrl(CHECK_STORY_ID), { headers: pageHeaders() });
  if (!response.ok) throw new Error(`HTTP ${response.status} checking the Chyoa session`);
  if (isLoginPage(await response.text())) return { status: "expired" };
  return { status: "ok" };
}
