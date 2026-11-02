import { afterAll, describe, expect, test } from "bun:test";
import { cookieHeader, jarFetch, jarOf, newJar, parseSetCookie, serializeJar } from "../src/account.ts";

// A site that logs in with a POST, sets its cookie on the redirect, and
// rotates the cookie on every logged-in page.
const site = Bun.serve({
  port: 0,
  fetch(request) {
    const url = new URL(request.url);
    if (url.pathname === "/login" && request.method === "POST") {
      return new Response(null, {
        status: 302,
        headers: { Location: "/home", "Set-Cookie": "sid=one; Path=/; HttpOnly; Expires=Wed, 21 Oct 2099 07:28:00 GMT" },
      });
    }
    if (url.pathname === "/home") {
      const cookie = request.headers.get("cookie") ?? "";
      if (!cookie.includes("sid=")) return new Response("login please", { status: 403 });
      return new Response(`hello ${request.method}`, { headers: { "Set-Cookie": "sid=two; Path=/" } });
    }
    return new Response("not found", { status: 404 });
  },
});

afterAll(() => site.stop());

describe("cookie jars", () => {
  test("a login keeps the cookie set on its redirect and follows as GET", async () => {
    const jar = newJar();
    const response = await jarFetch(jar, `${site.url}login`, { method: "POST", body: "user=me" });
    expect(await response.text()).toBe("hello GET");
    expect(jar.cookies.map((cookie) => `${cookie.name}=${cookie.value}`)).toEqual(["sid=two"]);
    expect(jarOf(serializeJar(jar))).toEqual(jar);
  });

  test("cookies go only where their domain, path and scheme allow", () => {
    const jar = jarOf(
      JSON.stringify({
        cookies: [
          { name: "a", value: "1", domain: ".site.test", path: "/" },
          { name: "b", value: "2", domain: "other.test", path: "/" },
          { name: "c", value: "3", domain: "site.test", path: "/members" },
          { name: "d", value: "4", domain: "site.test", secure: true },
          { name: "e", value: "5", domain: "site.test", expires: 1 },
        ],
      }),
    );
    expect(cookieHeader(jar, "http://www.site.test/")).toBe("a=1");
    expect(cookieHeader(jar, "https://site.test/members/x")).toBe("a=1; c=3; d=4");
  });

  test("Set-Cookie attributes are read", () => {
    const cookie = parseSetCookie("token=x=y; Domain=.site.test; Path=/app; Secure; Max-Age=60", "https://www.site.test/");
    expect(cookie?.value).toBe("x=y");
    expect(cookie?.domain).toBe("site.test");
    expect(cookie?.path).toBe("/app");
    expect(cookie?.secure).toBe(true);
    expect(cookie?.expires).toBeGreaterThan(Date.now() / 1000);
  });

  test("an unreadable session is an empty jar", () => {
    expect(jarOf("cookie=abc")).toEqual({ cookies: [] });
  });
});
