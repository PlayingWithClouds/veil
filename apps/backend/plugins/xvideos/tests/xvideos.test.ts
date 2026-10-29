import { test, expect, describe } from "bun:test";
import { readFileSync } from "fs";
import { jsonItem, jsonVideoList, parseCardsHtml } from "../src/cards.ts";
import { parseComments } from "../src/comments.ts";
import { canonicalUrl, countFromText, decodeEntities, secondsFromLabel, siteOf } from "../src/http.ts";
import { profilePath } from "../src/scenes.ts";
import { parseVideoPage, playerSettings, secondsFromIsoDuration } from "../src/scrape.ts";

/** Reads a saved page or answer from the fixtures folder. */
function fixture(name: string): string {
  return readFileSync(new URL(`./fixtures/${name}`, import.meta.url), "utf8");
}

describe("urls and labels", () => {
  test("tells the sites apart and reads them on their English www host", () => {
    expect(siteOf("https://de.xvideos.com/video.abc/x")).toBe("xvideos");
    expect(siteOf("https://www.xnxx.com/video-abc/x")).toBe("xnxx");
    expect(siteOf("https://example.com/")).toBeUndefined();
    expect(canonicalUrl("https://de.xvideos.com/video.abc/x?a=1")).toBe("https://www.xvideos.com/video.abc/x?a=1");
    expect(canonicalUrl("/video-abc/x", "xnxx")).toBe("https://www.xnxx.com/video-abc/x");
  });

  test("reads counts and runtimes", () => {
    expect(countFromText("931.8k")).toBe(931800);
    expect(countFromText("4.3M")).toBe(4300000);
    expect(countFromText("1,107")).toBe(1107);
    expect(secondsFromLabel("5 min")).toBe(300);
    expect(secondsFromLabel("8min")).toBe(480);
    expect(secondsFromLabel("1 h 12 min")).toBe(4320);
    expect(secondsFromIsoDuration("PT00H05M44S")).toBe(344);
  });

  test("decodes numeric and named entities", () => {
    expect(decodeEntities("Pool Guy verf&uuml;hrt &amp; &#039;more&#039; &hellip;")).toBe("Pool Guy verführt & 'more' …");
    expect(decodeEntities("&unknown;")).toBe("&unknown;");
  });

  test("maps listable pages to their profile", () => {
    expect(profilePath("/pornstars/nicole_aniston")).toBe("/pornstars/nicole_aniston");
    expect(profilePath("/profiles/18club/videos")).toBe("/profiles/18club");
    expect(profilePath("/bubblebooties")).toBe("/channels/bubblebooties");
    expect(profilePath("/video.abc/x")).toBeUndefined();
  });
});

describe("listing cards", () => {
  const cards = parseCardsHtml(fixture("search.html"), "xvideos");

  test("reads every video card with its details", () => {
    expect(cards.length).toBe(24);
    expect(cards[0]).toEqual({
      title: "Blonde Babe Nicole Aniston",
      media_type: "scene",
      source_url: "https://www.xvideos.com/video.hofbtbke2c2/blonde_babe_nicole_aniston",
      external_id: "xvideos-hofbtbke2c2",
      poster_path: "https://thumb-cdn77.xvideos-cdn.com/a7c798c9-d259-456d-8259-8378c554baae/0/xv_1_t.jpg",
      preview_video: "https://thumb-cdn77.xvideos-cdn.com/a7c798c9-d259-456d-8259-8378c554baae/0/preview.mp4",
      duration_seconds: 300,
      view_count: 931800,
      studio: {
        name: "Bangbros Bubble Butts",
        source_url: "https://www.xvideos.com/bubblebooties",
        external_id: "xvideos-uploader-bubblebooties",
      },
    });
  });

  test("reads the JSON listings of profiles and channels", () => {
    for (const name of ["pornstar-videos.json", "channel-videos.json"]) {
      const videos = jsonVideoList(JSON.parse(fixture(name)).videos);
      expect(videos.length).toBeGreaterThan(20);
      const item = jsonItem(videos[0], "xvideos")!;
      expect(item.external_id).toStartWith("xvideos-");
      expect(item.source_url).toStartWith("https://www.xvideos.com/video.");
      expect(item.title.length).toBeGreaterThan(0);
      expect(item.poster_path).toStartWith("https://");
    }
  });
});

describe("video page", () => {
  const url = "https://www.xvideos.com/video.hofbtbke2c2/blonde_babe_nicole_aniston";
  const page = parseVideoPage(fixture("video.html"), url, "xvideos");
  const scene = page.scene;

  test("reads the video's details", () => {
    expect(scene.external_id).toBe("xvideos-hofbtbke2c2");
    expect(scene.source_url).toBe(url);
    expect(scene.title).toBe("Blonde Babe Nicole Aniston");
    expect(scene.duration_seconds).toBe(344);
    expect(scene.date).toBe("2016-08-10T17:57:00.000Z");
    expect(scene.view_count).toBe(930794);
    expect(scene.rating).toBe(10);
    expect(scene.poster_path).toEndWith("/xv_1_t.jpg");
    expect(scene.preview_video).toEndWith("/preview.mp4");
    expect(scene.downloads).toEqual([{ label: "xvideos", url }]);
    expect(page.numericId).toBe("23048371");
  });

  test("credits the channel, the pornstar and the tags", () => {
    expect(scene.studio?.name).toBe("Bangbros Bubble Butts");
    expect(scene.performers).toEqual([
      {
        name: "Nicole Aniston",
        order: 0,
        source_url: "https://www.xvideos.com/pornstars/nicole_aniston",
        external_id: "xvideos-pornstars-nicole_aniston",
      },
    ]);
    expect(scene.tags).toEqual(["blonde", "babe", "blowjob", "bj", "big-tits", "nice-ass", "pawg"]);
  });

  test("lists the related videos without the video itself", () => {
    expect(scene.related!.length).toBe(40);
    expect(scene.related!.some((item) => item.external_id === scene.external_id)).toBe(false);
    expect(scene.related![0].studio?.name).toBe("Ass Parade");
    expect(scene.related![0].rating).toBe(9.9);
  });

  test("finds the streams the player was given", () => {
    const player = playerSettings(fixture("video.html"));
    expect(player.get("setVideoHLS")).toEndWith("/hls.m3u8");
    expect(player.get("setVideoUrlHigh")).toContain("video_360p.mp4");
  });

  test("reads xnxx pages the same way", () => {
    const xnxx = parseVideoPage(fixture("xnxx-video.html"), "https://www.xnxx.com/video-1iaga4ab/busty_blonde_teases", "xnxx");
    expect(xnxx.scene.external_id).toBe("xnxx-1iaga4ab");
    expect(xnxx.scene.title).toBe("Busty Blonde Teases");
    expect(xnxx.scene.performers!.map((performer) => performer.name)).toEqual(["Octavia Red", "Parker Ambrose"]);
    expect(xnxx.scene.tags).toContain("big ass");
    expect(xnxx.scene.studio?.name).toBe("blogporntv");
    expect(xnxx.scene.poster_path).toEndWith(".jpg");
    expect(xnxx.scene.related!.length).toBeGreaterThan(10);
    expect(xnxx.numericId).toBe("91186924");
  });
});

describe("comments", () => {
  test("reads the posts with author, date and likes", () => {
    const comments = parseComments(JSON.parse(fixture("comments.json")));
    expect(comments.length).toBe(1);
    expect(comments[0].text).toBe("I wanna suck a dick like she does!");
    expect(comments[0].author).toBe("Natasha1596");
    expect(comments[0].likes).toBe(0);
    expect(comments[0].date).toBeDefined();
  });

  test("puts replies after the comment they answer", () => {
    const comments = parseComments({
      posts: {
        ids: ["1", "2"],
        posts: {
          "1": { name: "a", message: "who is she?", replies: { ids: ["3"], posts: { "3": { name: "b", message: "Nicole<br />Aniston" } } } },
          "2": { name: "c", message: "<b>hot</b>" },
        },
      },
    });
    expect(comments).toEqual([
      { text: "who is she?", author: "a" },
      { text: "Nicole\nAniston", author: "b", reply_to: 0 },
      { text: "hot", author: "c" },
    ]);
  });
});
