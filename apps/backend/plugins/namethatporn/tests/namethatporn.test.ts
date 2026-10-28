import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parseAnswerNames, parsePosts } from "../src/parse.ts";

/** Loads a saved page or picture from the fixtures folder. */
function loadFixture(path: string): Buffer {
  return readFileSync(new URL(path, import.meta.url));
}

describe("parsePosts", () => {
  const posts = parsePosts(loadFixture("./fixtures/results.html").toString("utf8"));

  test("reads every result in the site's order", () => {
    expect(posts.length).toBe(68);
    expect(posts[0].url).toBe("https://namethatporn.com/post/439897-who-is-she.html");
  });

  test("reads title, thumbnail and whether the post is answered", () => {
    expect(posts[0]).toEqual({
      title: "Who is she?",
      url: "https://namethatporn.com/post/439897-who-is-she.html",
      thumbnail_url: "https://namethatporn.com/media/squares/2016/10/03/K4FIG948SHWAY_who-is-she.jpg",
      solved: true,
      names: [],
    });
    expect(posts[1].solved).toBe(false);
    expect(posts[2].title).toBe("What's the name of this porn actor?");
    expect(posts.filter((post) => post.solved).length).toBe(51);
  });

  test("a page without results has no posts", () => {
    expect(parsePosts("<main><div id='nsw_tls'>Found total of 0 results</div></main>")).toEqual([]);
  });
});

describe("parseAnswerNames", () => {
  test("reads the accepted answer's name tags only", () => {
    expect(parseAnswerNames(loadFixture("./fixtures/post.html").toString("utf8"))).toEqual(["Siri"]);
  });

  test("an unsolved post names nobody", () => {
    expect(parseAnswerNames("<div id='answer_1'><a class='name_syntax'><span class='name'>Siri</span></a></div>")).toEqual([]);
  });
});

// No live test here: Cloudflare turns Bun's fetch away (403) while Go's, which
// runs the plugin in the backend, gets through; TestEmbeddedBundlesLive covers it.
