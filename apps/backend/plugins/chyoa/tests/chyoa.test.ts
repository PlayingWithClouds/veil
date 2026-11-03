import { afterEach, describe, expect, test } from "bun:test";
import { readFileSync } from "fs";
import { NeedsLoginError } from "@playingwithclouds/veil-sdk";
import { parseStoryCards, sourceOf, sourceOfUrl, storyList } from "../src/listing.ts";
import { checkAccount, csrfTokenOf, login, loginErrorOf } from "../src/login.ts";
import { chapterIdOf, isLoginPage, storyIdOf, storyUrl } from "../src/site.ts";
import { chapterFromPage, findStory, readChapter, storyFromPage, writtenDate } from "../src/story.ts";
import { countWords, plainText } from "../src/text.ts";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/** Loads a saved page from the fixtures folder. */
function loadFixture(fileName: string): string {
  return readFileSync(new URL(`./fixtures/${fileName}`, import.meta.url), "utf8");
}

/** A page answer. */
function page(html: string, status = 200): Response {
  return new Response(html, { status, headers: { "Content-Type": "text/html" } });
}

/** A redirect answer, optionally setting cookies. */
function redirect(location: string, cookies: string[] = []): Response {
  const headers = new Headers({ Location: location });
  for (const cookie of cookies) headers.append("Set-Cookie", cookie);
  return new Response(null, { status: 302, headers });
}

/** One request the mocked site saw. */
interface SeenRequest {
  url: string;
  method: string;
  cookie: string;
  body: string;
}

/** Replaces fetch with a site answering by its handler, and returns the requests seen. */
function mockSite(handler: (request: SeenRequest) => Response): SeenRequest[] {
  const seen: SeenRequest[] = [];
  globalThis.fetch = (async (input: string | URL | Request, init?: RequestInit) => {
    const headers = (init?.headers ?? {}) as Record<string, string>;
    const request: SeenRequest = {
      url: String(input),
      method: init?.method ?? "GET",
      cookie: headers.Cookie ?? "",
      body: String(init?.body ?? ""),
    };
    seen.push(request);
    return handler(request);
  }) as typeof fetch;
  return seen;
}

describe("addresses", () => {
  test("story and chapter ids come out of their addresses", () => {
    expect(storyIdOf("https://chyoa.com/story/Attaining-Tenure.41407")).toBe("41407");
    expect(storyIdOf("https://chyoa.com/story/Mr.-X.123?a=1")).toBe("123");
    expect(storyIdOf("https://chyoa.com/story/new")).toBeUndefined();
    expect(storyIdOf("https://chyoa.com/chapter/Intro.5")).toBeUndefined();
    expect(chapterIdOf("https://chyoa.com/chapter/The-Principal%27s-Office.1071487")).toBe("1071487");
    expect(chapterIdOf("https://chyoa.com/chapter/Pt.-2.77/bookmark")).toBe("77");
    expect(chapterIdOf("https://chyoa.com/chapter/1071487")).toBeUndefined();
    expect(chapterIdOf("https://example.com/chapter/a.1")).toBeUndefined();
  });

  test("a story is addressed by id, the site ignoring the slug", () => {
    expect(storyUrl("41407")).toBe("https://chyoa.com/story/x.41407");
  });
});

describe("text", () => {
  test("paragraphs, headings and rules become blank lines; pictures vanish", () => {
    const markup = `<h2>Warning</h2>\n<p>One &quot;quoted&quot; line<br />still one</p>\n<p class="rd-story-media"><img src="x.png" alt="" /></p>\n<hr />\n<p>Two&nbsp;words &amp; more</p>`;
    expect(plainText(markup)).toBe('Warning\n\nOne "quoted" line\nstill one\n\nTwo words & more');
  });

  test("words are counted", () => {
    expect(countWords("one two\n\nthree")).toBe(3);
    expect(countWords("")).toBe(0);
  });

  test("a written date becomes an ISO date", () => {
    expect(writtenDate("Nov 2, 2026 by someone")).toBe("2026-11-02");
    expect(writtenDate("Sep 27, 2026")).toBe("2026-09-27");
    expect(writtenDate("yesterday")).toBeUndefined();
  });
});

describe("listing pages", () => {
  const items = parseStoryCards(loadFixture("listing-new.html"));

  test("every story card becomes an item", () => {
    expect(items.length).toBe(3);
    expect(items[0]).toEqual({
      title: "Fucking",
      media_type: "story",
      source_url: "https://chyoa.com/story/Fucking.76880",
      external_id: "chyoa-76880",
      poster_path: "https://chyoa.com/data/covers/s/76/76880.jpg?1793408741",
      date: "2026-11-02",
      view_count: 50,
      studio: {
        name: "Somefucker123",
        external_id: "chyoa-user-somefucker123",
        source_url: "https://chyoa.com/user/Somefucker123",
      },
    });
  });

  test("a card's studio is the identity the story page credits as performer", () => {
    const story = storyFromPage(loadFixture("story-ending.html"), "https://chyoa.com/story/x.76880");
    expect(story?.performers?.[0]).toEqual(items[0].studio);
    expect(items.every((item) => item.studio?.external_id?.startsWith("chyoa-user-"))).toBe(true);
  });

  test("a card without a cover has no poster, and view counts lose their thousands separators", () => {
    expect(items[1].poster_path).toBeUndefined();
    expect(items[2].view_count).toBe(1145);
  });

  test("no query and no URL lists the newest stories", () => {
    expect(sourceOf({}).pageUrl(2)).toBe("https://chyoa.com/new-sex-stories/page/2");
  });

  test("a query searches stories", () => {
    expect(sourceOf({ query: "big  tits" }).pageUrl(3)).toBe("https://chyoa.com/search?q=big%20tits&type=story&page=3");
  });

  test("category, ranking and search addresses are recognised", () => {
    expect(sourceOfUrl("https://chyoa.com/category/fetish")?.pageUrl(2)).toBe("https://chyoa.com/category/fetish?page=2");
    expect(sourceOfUrl("https://chyoa.com/most-liked-sex-stories")?.pageUrl(2)).toBe(
      "https://chyoa.com/most-liked-sex-stories/page/2",
    );
    expect(sourceOfUrl("https://chyoa.com/search?q=mother&type=story&page=4")?.pageUrl(1)).toBe(
      "https://chyoa.com/search?q=mother&type=story&page=1",
    );
    expect(sourceOfUrl("https://chyoa.com/story/a.1")).toBeUndefined();
  });

  test("limit and offset slice across pages and stop at a short page", async () => {
    const html = loadFixture("listing-new.html");
    const requested: string[] = [];
    mockSite((request) => {
      requested.push(request.url);
      return page(html);
    });
    const result = await storyList({ offset: 1, limit: 5 });
    expect(requested).toEqual(["https://chyoa.com/new-sex-stories/page/1"]);
    expect(result.items.map((item) => item.external_id)).toEqual(["chyoa-76916", "chyoa-76915"]);
  });
});

describe("story pages", () => {
  test("a branching story gets its first chapter and the first options as children", () => {
    const story = storyFromPage(loadFixture("story-branching.html"), "https://chyoa.com/story/x.41407");
    expect(story?.external_id).toBe("chyoa-41407");
    expect(story?.title).toBe("Attaining Tenure");
    expect(story?.source_url).toBe("https://chyoa.com/story/Attaining-Tenure.41407");
    expect(story?.details).toStartWith("Miss Leah Williams is on her final year");
    expect(story?.date).toBe("2022-06-01");
    expect(story?.performers?.[0].name).toBe("sexybjgal69");
    expect(story?.tags?.slice(0, 3)).toEqual(["Cheating Spouses", "blackmail", "non-consent"]);
    expect(story?.cover_path).toBe("https://chyoa.com/data/covers/l/41/41407.jpg?1641683395");

    const [first, ...options] = story?.chapters ?? [];
    expect(first.url).toBe("https://chyoa.com/chapter/Introduction.1071482");
    expect(first.parent_url).toBeUndefined();
    expect(first.text).toContain("Leah Williams sat at her desk");
    expect(first.text).not.toContain("<p>");
    expect(first.word_count).toBeGreaterThan(100);
    expect(options.map((option) => option.choice)).toEqual([
      "The Principal's Office",
      "Character Creation****Game Info***** Read First****",
      "Road Map, Updates, Feedback, Requests",
    ]);
    for (const option of options) {
      expect(option.parent_url).toBe(first.url);
      expect(option.text).toBeUndefined();
    }
    expect(options[0].url).toBe("https://chyoa.com/chapter/The-Principal%27s-Office.1071487");
  });

  test("a story whose only chapter ends has no children and no complete flag", () => {
    const story = storyFromPage(loadFixture("story-ending.html"), "https://chyoa.com/story/x.76880");
    expect(story?.chapters?.length).toBe(1);
    expect(story?.complete).toBeUndefined();
    expect(story?.tags?.[0]).toBe("Fetish");
  });

  test("a page that is no story gives nothing", () => {
    expect(storyFromPage("<html><body>nothing</body></html>", "https://chyoa.com/story/x.1")).toBeUndefined();
  });

  test("a deeper chapter reads its text and the chapters its options lead to", () => {
    const chapter = chapterFromPage(loadFixture("chapter-deep.html"), "https://chyoa.com/chapter/The-Principal%27s-Office.1071487");
    expect(chapter?.title).toBe("The Principal's Office");
    expect(chapter?.text).toContain("Leah headed toward the office");
    expect(chapter?.choices).toEqual([
      {
        title: "Refuse The Principal's Quid Pro Quo",
        url: "https://chyoa.com/chapter/Refuse-The-Principal%27s-Quid-Pro-Quo.1071521",
        choice: "Refuse The Principal's Quid Pro Quo",
      },
      {
        title: "Accept the Quid Pro Quo",
        url: "https://chyoa.com/chapter/Accept-the-Quid-Pro-Quo.1071522",
        choice: "Accept the Quid Pro Quo",
      },
      {
        title: "Character Creation",
        url: "https://chyoa.com/chapter/Character-Creation.1447285",
        choice: "Character Creation",
      },
    ]);
  });

  test("an ending has no choices", () => {
    const chapter = chapterFromPage(loadFixture("story-ending.html"), "https://chyoa.com/chapter/Introduction.1917360");
    expect(chapter?.choices).toBeUndefined();
    expect(chapter?.text).toContain("HAPPY HALLOWEEENNNNNNN");
  });
});

describe("reading with and without a login", () => {
  test("find fetches the story by id and keeps the account's cookies", async () => {
    const seen = mockSite(() => page(loadFixture("story-branching.html")));
    const results = await findStory("https://chyoa.com/story/Attaining-Tenure.41407");
    expect(seen.map((request) => request.url)).toEqual(["https://chyoa.com/story/x.41407"]);
    expect(results[0].type).toBe("story");
  });

  test("find given a chapter address goes through the chapter's Start Over link", async () => {
    const seen = mockSite((request) => {
      if (request.url.includes("/chapter/")) return page(loadFixture("chapter-deep.html"));
      return page(loadFixture("story-branching.html"));
    });
    await findStory("https://chyoa.com/chapter/The-Principal%27s-Office.1071487");
    expect(seen.map((request) => request.url)).toEqual([
      "https://chyoa.com/chapter/The-Principal%27s-Office.1071487",
      "https://chyoa.com/story/Attaining-Tenure.41407",
    ]);
  });

  test("a chapter read returns its text and choices", async () => {
    mockSite(() => page(loadFixture("chapter-deep.html")));
    const chapter = await readChapter("https://chyoa.com/chapter/The-Principal%27s-Office.1071487");
    expect(chapter.choices?.length).toBe(3);
  });

  test("a redirect to the login page asks for a login", async () => {
    mockSite((request) => {
      if (request.url.endsWith("/auth/login")) return page(loadFixture("login-page.html"));
      return redirect("https://chyoa.com/auth/login");
    });
    await expect(readChapter("https://chyoa.com/chapter/The-Principal%27s-Office.1071487")).rejects.toBeInstanceOf(NeedsLoginError);
    await expect(findStory("https://chyoa.com/story/Attaining-Tenure.41407")).rejects.toBeInstanceOf(NeedsLoginError);
  });

  test("a login form in place of the page asks for a login even when the URL isn't known", async () => {
    mockSite(() => page(loadFixture("login-page.html")));
    await expect(readChapter("https://chyoa.com/chapter/A.1")).rejects.toBeInstanceOf(NeedsLoginError);
  });

  test("a deleted chapter is an error, not a login", async () => {
    mockSite(() => page("<html><h1>This branch was never written</h1></html>", 404));
    const failure = await readChapter("https://chyoa.com/chapter/A.1").catch((error: Error) => error);
    expect(failure).toBeInstanceOf(Error);
    expect(failure).not.toBeInstanceOf(NeedsLoginError);
  });

  test("addresses of other sites are refused", async () => {
    await expect(readChapter("https://example.com/chapter/a.1")).rejects.toThrow("Not a chyoa.com chapter address");
    expect(await findStory("https://example.com/story/a.1")).toEqual([]);
  });
});

describe("login", () => {
  test("the login page's token and the site's refusal are read", () => {
    expect(isLoginPage(loadFixture("login-page.html"))).toBe(true);
    expect(isLoginPage(loadFixture("chapter-deep.html"))).toBe(false);
    expect(csrfTokenOf(loadFixture("login-page.html"))).toBe("CSRF_TOKEN_PLACEHOLDER");
    expect(csrfTokenOf("<html></html>")).toBeUndefined();
    expect(loginErrorOf(loadFixture("login-failed.html"))).toBe("These credentials do not match our records.");
    expect(loginErrorOf(loadFixture("login-page.html"))).toBe("Chyoa did not accept this username and password");
  });

  /** The site as a login sees it: a form to fetch, a POST that redirects home or back. */
  function loginSite(accepts: boolean): SeenRequest[] {
    let posted = false;
    return mockSite((request) => {
      if (request.method === "POST") {
        posted = true;
        if (!accepts) return redirect("https://chyoa.com/auth/login", ["laravel_session=refused; path=/"]);
        return redirect("https://chyoa.com/", ["laravel_session=loggedin; path=/", "remember_web=abc; path=/; max-age=99999999"]);
      }
      if (request.url === "https://chyoa.com/") return page('<html><body><a href="/auth/logout">Log out</a></body></html>');
      if (posted) return page(loadFixture("login-failed.html"));
      return new Response(loadFixture("login-page.html"), {
        headers: { "Set-Cookie": "laravel_session=anonymous; path=/; httponly" },
      });
    });
  }

  test("a login posts the token, the credentials and the session cookie, and keeps the cookies set on the way", async () => {
    const seen = loginSite(true);
    const result = await login({ username: " alice ", password: "p&ss word" });
    const post = seen.find((request) => request.method === "POST");
    expect(post?.cookie).toBe("laravel_session=anonymous");
    const form = new URLSearchParams(post?.body);
    expect(form.get("_token")).toBe("CSRF_TOKEN_PLACEHOLDER");
    expect(form.get("username")).toBe("alice");
    expect(form.get("password")).toBe("p&ss word");
    expect(form.get("remember")).toBe("1");

    expect(result.status).toBe("ok");
    expect(result.label).toBe("alice");
    const cookies = JSON.parse(result.session ?? "{}").cookies as { name: string; value: string; domain: string }[];
    expect(cookies.find((cookie) => cookie.name === "laravel_session")?.value).toBe("loggedin");
    expect(cookies.find((cookie) => cookie.name === "remember_web")?.domain).toBe("chyoa.com");
  });

  test("a refused login fails with the site's message", async () => {
    loginSite(false);
    const result = await login({ username: "alice", password: "wrong" });
    expect(result.status).toBe("failed");
    expect(result.error).toBe("These credentials do not match our records.");
    expect(result.session).toBeUndefined();
  });

  test("missing credentials fail without a request", async () => {
    const seen = loginSite(true);
    const result = await login({ username: "alice" });
    expect(result.status).toBe("failed");
    expect(seen.length).toBe(0);
  });

  test("a login page without a form fails", async () => {
    mockSite(() => page("<html><body>Maintenance</body></html>"));
    const result = await login({ username: "alice", password: "x" });
    expect(result.status).toBe("failed");
  });
});

describe("account check", () => {
  const session = JSON.stringify({
    cookies: [{ name: "laravel_session", value: "abc", domain: "chyoa.com", path: "/" }],
  });

  test("a session that reads a story is ok and sends its cookies", async () => {
    const seen = mockSite(() => page(loadFixture("story-branching.html")));
    expect(await checkAccount(session)).toEqual({ status: "ok" });
    expect(seen[0].cookie).toBe("laravel_session=abc");
  });

  test("a session sent to the login page has expired", async () => {
    mockSite((request) => {
      if (request.url.endsWith("/auth/login")) return page(loadFixture("login-page.html"));
      return redirect("https://chyoa.com/auth/login");
    });
    expect(await checkAccount(session)).toEqual({ status: "expired" });
  });

  test("a site error is an error, not an expired session", async () => {
    mockSite(() => page("down", 503));
    await expect(checkAccount(session)).rejects.toThrow("HTTP 503");
  });
});
