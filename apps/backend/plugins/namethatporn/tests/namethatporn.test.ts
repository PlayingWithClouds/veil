import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { parseAnswerNames, parsePosts } from "../src/parse.ts";
import { parseThread } from "../src/thread.ts";

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

describe("parseThread", () => {
  /** Parses a saved thread page. */
  function thread(fileName: string, url: string) {
    return parseThread(loadFixture(`./fixtures/${fileName}`).toString("utf8"), url);
  }

  test("reads a video post, its accepted answer once, and its unwrapped link", () => {
    const post = thread("post-video.html", "https://namethatporn.com/post/1863105.html");
    expect(post.title).toBe("What's the title for this letsdoeit squirting gif ad?");
    expect(post.description).toBe("gifpigeon.com");
    expect(post.video_url).toEndWith(".mp4");
    expect(post.tags).toContain("letsdoeit");
    expect(post.solved).toBe(true);
    const accepted = post.comments.filter((comment) => comment.accepted);
    expect(accepted.length).toBe(1);
    expect(accepted[0].names).toEqual(["Sara Diamante"]);
    expect(accepted[0].titles).toEqual(["prankster scores roomie pussy"]);
    expect(accepted[0].links).toEqual(["https://www.letsdoeit.com/scene/11469265/prankster-scores-roomie-pussy"]);
    expect(accepted[0].text).toContain("Here is link:");
    expect(new Set(post.comments.map((comment) => comment.id)).size).toBe(post.comments.length);
    expect(post.image_url).toEndWith(".jpg");
  });

  test("keeps cards of other threads out of the text, as references", () => {
    const post = thread("post-video.html", "https://namethatporn.com/post/1863105.html");
    const withCards = post.comments.find((comment) => comment.posts.length > 0)!;
    expect(withCards.posts[0]).toEqual({
      title: "Full video plz",
      url: "https://namethatporn.com/post/1857589-full-video-plz.html",
      thumbnail_url: "https://namethatporn.com/media/tinies/2026/06/06/HVN2PHKBHERFA_full-video-plz.jpg",
    });
    expect(withCards.text).not.toContain("Full video plz");
    expect(withCards.links).toEqual([]);
  });

  test("reads every comment with author, likes, pictures and bare addresses", () => {
    const post = thread("post-thread.html", "https://namethatporn.com/post/496735.html");
    expect(post.image_url).toEndWith(".jpg");
    expect(post.comments.length).toBe(4);
    const [photos, bare, , nameOnly] = post.comments;
    expect(nameOnly.names).toEqual(["Jasmine James"]);
    expect(photos.accepted).toBe(true);
    expect(photos.author).toBe("Deleted User");
    expect(photos.likes).toBe(2);
    expect(photos.date).toBe("2017-07-01T00:36:17.000Z");
    expect(photos.images.length).toBe(2);
    expect(photos.links).toEqual([]);
    expect(photos.text).toStartWith("Just a photoshoot\n\nShe is Jasmine James");
    expect(bare.links).toEqual(["https://www.pornburst.xxx/videos/busty-nurse-fucks-a-huge-cock/"]);
    expect(bare.date).toBeDefined();
  });

  test("unwraps the site's redirect around a tube link", () => {
    const post = thread("post-link.html", "https://namethatporn.com/post/913519.html");
    const links = post.comments.flatMap((comment) => comment.links);
    expect(links).toContain("https://pornhub.com/view_video.php?viewkey=ph563f78988d71d");
  });
});
