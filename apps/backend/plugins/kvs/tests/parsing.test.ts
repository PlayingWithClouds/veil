import { test, expect, describe } from "bun:test";
import { parse } from "node-html-parser";
import {
  contentKinds,
  enabledSites,
  siteByKey,
  siteForUrl,
  sitesCarrying,
  supportedDomains,
  SITES,
  type SiteConfig,
} from "../src/sites.ts";
import { canonicalUrl, entitySlug, pathPageUrl, searchUrl, videoKey } from "../src/urls.ts";
import { asyncPageUrl, parseBlock } from "../src/crawl.ts";
import { entityName, parseDurationText, parseIsoDuration, parseMinuteBadge, withoutRankBadge } from "../src/text.ts";
import { extractSources, parseFlashvars } from "../src/media.ts";
import { parseListingCards, parseVideoCards } from "../src/listing.ts";
import { MODELS, TAGS, parseCatalogEntries, splitCount } from "../src/catalog.ts";
import { interleave } from "../src/fanout.ts";
import { mimeTypeOf, requestedQuality } from "../src/resolve.ts";
import { camPlatformStudio, withoutSiteSuffix } from "../src/scrape.ts";
import type { DiscoveredItem } from "@playingwithclouds/veil-sdk";

/** The configured site with this key; fails the test when missing. */
function site(key: string): SiteConfig {
  const found = siteByKey(key);
  if (!found) throw new Error(`no site ${key}`);
  return found;
}

describe("site config", () => {
  test("keys are unique and every domain routes back to its site", () => {
    const keys = SITES.map((entry) => entry.key);
    expect(new Set(keys).size).toBe(keys.length);
    for (const domain of supportedDomains()) {
      expect(siteForUrl(`https://${domain}/`)).toBeDefined();
    }
  });

  test("www, language and mirror subdomains map to the site", () => {
    expect(siteForUrl("https://de.pornhat.com/video/x/")?.key).toBe("pornhat");
    expect(siteForUrl("https://w11.mygoodporn.tv/video/1/x/")?.key).toBe("mygoodporn");
    expect(siteForUrl("https://ok.xxx/video/1/")?.key).toBe("okxxx");
    expect(siteForUrl("https://example.com/video/1/")).toBeUndefined();
  });

  test("a search for cam shows spans only the cam archives", () => {
    const camSites = sitesCarrying("cam-shows").map((entry) => entry.key);
    expect(camSites.sort()).toEqual(["camstreams", "cloudbate"]);
    expect(sitesCarrying("videos").some((entry) => entry.key === "cloudbate")).toBe(false);
    expect(sitesCarrying(undefined).length).toBe(enabledSites().length);
    expect(contentKinds().sort()).toEqual(["cam-shows", "videos"]);
  });
});

describe("urls", () => {
  test("videoKey follows each site's URL scheme", () => {
    expect(videoKey(site("okxxx"), "/video/791051/")).toBe("791051");
    expect(videoKey(site("pornhat"), "https://www.pornhat.com/video/big-ass-creampie/")).toBe("big-ass-creampie");
    expect(videoKey(site("xcafe"), "https://xcafe.com/97539/")).toBe("97539");
    expect(videoKey(site("xcafe"), "https://xcafe.com/videos/blonde/")).toBe("");
    expect(videoKey(site("xcum"), "https://xcum.com/v/311019/")).toBe("311019");
    expect(videoKey(site("porntrex"), "https://www.porntrex.com/video/3356405/some-slug")).toBe("3356405");
    expect(videoKey(site("inxxx"), "/v/some-title.xxx-video")).toBe("some-title");
    // Another site's link is not this site's video.
    expect(videoKey(site("okxxx"), "https://www.pornhat.com/video/123/")).toBe("");
  });

  test("canonicalUrl moves pages onto the origin without query or fragment", () => {
    expect(canonicalUrl(site("pornhat"), "https://de.pornhat.com/video/x/?utm=1#720p")).toBe(
      "https://www.pornhat.com/video/x/"
    );
  });

  test("entitySlug skips pagination and other prefixes", () => {
    const anysex = site("anysex");
    expect(entitySlug(anysex, "https://anysex.com/videos/categories/hardcore/", anysex.tagPrefixes)).toBe("hardcore");
    expect(entitySlug(site("okxxx"), "/models/2/", ["/models/"])).toBe("");
    expect(entitySlug(site("okxxx"), "/tags/hd/", ["/models/"])).toBe("");
  });

  test("searchUrl fills the dash-joined or form-encoded query", () => {
    expect(searchUrl(site("okxxx"), " big  tits ")).toBe("https://ok.xxx/search/big-tits/");
    expect(searchUrl(site("anysex"), "big tits")).toBe("https://anysex.com/search/?q=big+tits");
  });

  test("pathPageUrl appends the page before the query", () => {
    expect(pathPageUrl("https://anysex.com/search/?q=blonde", 2)).toBe("https://anysex.com/search/2/?q=blonde");
    expect(pathPageUrl("https://ok.xxx/", 3)).toBe("https://ok.xxx/3/");
    expect(pathPageUrl("https://xxxi.porn/new-porn", 2)).toBe("https://xxxi.porn/new-porn/2/");
    expect(pathPageUrl("https://ok.xxx/", 1)).toBe("https://ok.xxx/");
  });
});

describe("async pagination", () => {
  test("parseBlock finds the page keys and keeps the rest", () => {
    const block = parseBlock("list_videos_search", "q:big%20tits;category_ids:;sort_by:;from_videos+from_albums:02");
    expect(block?.pageKeys).toEqual(["from_videos", "from_albums"]);
    expect(block?.parameters).toEqual([
      ["q", "big tits"],
      ["category_ids", ""],
      ["sort_by", ""],
    ]);
  });

  test("filters named *_from are no page keys", () => {
    expect(parseBlock("list", "sort_by:post_date;post_date_from:;duration_from:")).toBeUndefined();
  });

  test("asyncPageUrl requests the block with the page number", () => {
    const block = parseBlock("list_videos_common_videos_list", "sort_by:post_date;from:02");
    if (!block) throw new Error("no block");
    const url = new URL(asyncPageUrl("https://anyporn.com/categories/teens/", block, 3));
    expect(url.searchParams.get("mode")).toBe("async");
    expect(url.searchParams.get("function")).toBe("get_block");
    expect(url.searchParams.get("block_id")).toBe("list_videos_common_videos_list");
    expect(url.searchParams.get("sort_by")).toBe("post_date");
    expect(url.searchParams.get("from")).toBe("3");
  });
});

describe("text", () => {
  test("parseDurationText reads clock and unit forms", () => {
    expect(parseDurationText("14:16")).toBe(856);
    expect(parseDurationText("1:02:03")).toBe(3723);
    expect(parseDurationText("5m:59s")).toBe(359);
    expect(parseDurationText("Duration: 46min 55sec")).toBe(2815);
    expect(parseDurationText("no runtime")).toBe(0);
  });

  test("parseIsoDuration reads JSON-LD and itemprop durations", () => {
    expect(parseIsoDuration("PT00H12M01S")).toBe(721);
    expect(parseIsoDuration("PT0H7M17S")).toBe(437);
    expect(parseIsoDuration("T8M01S")).toBe(481);
    expect(parseIsoDuration("")).toBe(0);
  });

  test("withoutRankBadge drops a leading #N only", () => {
    expect(withoutRankBadge("#1Dainty Wilder")).toBe("Dainty Wilder");
    expect(withoutRankBadge("#Hashtag")).toBe("#Hashtag");
  });

  test("entityName drops counters and prefers name elements", () => {
    const root = parse(`
      <a id="a" href="/channels/adult-time/"><i></i>Adult Time<span class="count">1348</span></a>
      <a id="b" href="/models/tony/"><span class="name"><i></i> Tony Rubino </span><span class="button-info"> 2 </span></a>
      <a id="c" href="/sites/x/"><span class="text">Glowing Desire</span> <span>11</span></a>
      <a id="d" href="/tags/x/">-</a>
      <a id="e" href="/onlyfans-creators/dainty/"><span>#1</span><div class="text-truncate">Dainty Wilder</div></a>`);
    expect(entityName(root.querySelector("#a")!)).toBe("Adult Time");
    expect(entityName(root.querySelector("#b")!)).toBe("Tony Rubino");
    expect(entityName(root.querySelector("#c")!)).toBe("Glowing Desire");
    expect(entityName(root.querySelector("#d")!)).toBe("");
    expect(entityName(root.querySelector("#e")!)).toBe("Dainty Wilder");
  });

  test("withoutSiteSuffix strips the site name and SEO tails only", () => {
    expect(withoutSiteSuffix("Some Scene | Any Porn", site("anyporn"))).toBe("Some Scene");
    expect(withoutSiteSuffix("Some Scene - PornGO.com", site("porngo"))).toBe("Some Scene");
    expect(withoutSiteSuffix("A scene watch online | GiG.SEX", site("gigsex"))).toBe("A scene");
    expect(withoutSiteSuffix("Tennis Toes 2 - Flixxx", site("analdin"))).toBe("Tennis Toes 2 - Flixxx");
  });
});

describe("media", () => {
  const flashvarsPage = `
    <script>
      var t86b401a0ea = {
        video_id: '320793', video_title: 'Scene',
        timeline_screens_url: 'https://x/{time}.jpg',
        license_code: '$494000116297929',
        event_reporting2: 'https://yesporn.vip/get_file/5/0000000000000000000000000000000000/75000/75378/75378.mp4/?v-acctoken=pixel',
        video_url: 'function/0/https://yesporn.vip/get_file/5/7d6fd114c9e7289c81e78749465d532f735912fc1a/75000/75378/75378_trim.mp4/',
        video_url_text: 'SD',
        video_alt_url: 'https://yesporn.vip/get_file/5/aaaabbbbccccddddeeeeffff0000111122/75000/75378/75378_720p.mp4/?v-acctoken=abc',
        video_alt_url_text: 'HD'
      };
    </script>
    <script type="application/ld+json">{"contentUrl": "https://yesporn.vip/get_file/5/aaaabbbbccccddddeeeeffff0000111122/75000/75378/75378_720p.mp4/"}</script>`;

  test("parseFlashvars finds renamed objects with braces in values", () => {
    const flashvars = parseFlashvars(flashvarsPage);
    expect(flashvars.video_id).toBe("320793");
    expect(flashvars.timeline_screens_url).toBe("https://x/{time}.jpg");
    expect(flashvars.video_alt_url_text).toBe("HD");
  });

  test("extractSources decodes, labels and ranks flashvars sources, keeping their tokens", () => {
    const yesporn = site("yesporn");
    const sources = extractSources(yesporn, parse(flashvarsPage), flashvarsPage);
    expect(sources.map((source) => source.quality)).toEqual(["720p", "480p"]);
    expect(sources[0].url).toContain("?v-acctoken=abc");
    expect(sources[1].url).toBe(
      "https://yesporn.vip/get_file/5/4d84d12e7ff61595939c878cd4e72617735912fc1a/75000/75378/75378_trim.mp4/"
    );
  });

  test("extractSources reads <source> tags, script maps and skips previews and Auto", () => {
    const html = `
      <video>
        <source src="https://ok.xxx/get_file/13/7abb8c9d3cbc892f198ca7b2bce05bfb/791000/791051/791051_360p.mp4/" title="Auto">
        <source src="https://ok.xxx/get_file/13/7abb8c9d3cbc892f198ca7b2bce05bfb/791000/791051/791051_360p.mp4/" title="360p">
        <source src="https://ok.xxx/get_file/13/d3b555d9681dbc71a6c767d478afd87f/791000/791051/791051_720p.mp4/" title="720p">
      </video>
      <a data-preview="https://ok.xxx/get_file/13/2c556e9b9d4b3bde3cffd06c0076e346/790000/790376/790376_preview360p.mp4/"></a>
      <script>const sources = { 1080 : 'https://ok.xxx/get_file/13/11111111111111111111111111111111/791000/791051/791051.mp4/?br=9' };</script>`;
    const sources = extractSources(site("okxxx"), parse(html), html);
    expect(sources.map((source) => source.quality)).toEqual(["1080p", "720p", "360p"]);
  });

  test("extractSources accepts encrypted get_file paths", () => {
    const html = `<script>var flashvars = { license_code: '$467710015430571',
      video_url: 'https://fapgoat.com/get_file/0/jiolDu0p-Oo8HpcYyZSSnwl_vXLVE18t5DF3j43Xd-0Oxoj.mp4/?v-acctoken=x', video_url_text: '720p' };</script>`;
    const sources = extractSources(site("fapgoat"), parse(html), html);
    expect(sources).toHaveLength(1);
    expect(sources[0].quality).toBe("720p");
  });
});

describe("listing cards", () => {
  test("stock KVS cards: title, poster, duration, preview, channel", () => {
    const html = `
      <div id="list_videos_latest_videos_list_items">
        <div class="item">
          <a href="https://www.analdin.com/videos/1/first-video/" vthumb="https://www.analdin.com/get_file/1/c8b26fb08122db079105e9311243c4ea/1000/1/1_vthumb.mp4/">
            <div class="img"><img class="thumb" src="data:image/gif;base64,AAA" data-original="https://i.analdin.com/1.jpg"></div>
            <strong class="title"> First video </strong>
            <div class="duration">12:00</div>
          </a>
          <a href="https://www.analdin.com/channels/wtf/">wtf</a>
        </div>
        <div class="item">
          <a href="https://www.analdin.com/videos/2/second/" title="Second video"><img data-original="//i.analdin.com/2.jpg"></a>
        </div>
      </div>
      <div class="sidebar"><a href="https://www.analdin.com/videos/3/teaser/" title="Teaser">x</a></div>`;
    const items = parseListingCards(site("analdin"), parse(html));
    expect(items.map((item) => item.external_id)).toEqual(["kvs-analdin-1", "kvs-analdin-2"]);
    const [first, second] = items as [DiscoveredItem, DiscoveredItem];
    expect(first.title).toBe("First video");
    expect(first.poster_path).toBe("https://i.analdin.com/1.jpg");
    expect(first.duration_seconds).toBe(720);
    expect(first.preview_video).toContain("_vthumb.mp4");
    expect(first.studio).toEqual({
      name: "wtf",
      external_id: "kvs-analdin-channel-wtf",
      source_url: "https://www.analdin.com/channels/wtf/",
    });
    expect(second.title).toBe("Second video");
    expect(second.poster_path).toBe("https://i.analdin.com/2.jpg");
  });

  test("site title selector and bare runtime spans (xcum)", () => {
    const html = `
      <div class="thumbs">
        <div class="thumb"><a href="https://xcum.com/v/1/"><img alt="Performer A, Performer B" data-original="https://img/1.jpg">
          <span class="inf"><span>Real title</span><span></span><span>5:59</span><span>1</span></span></a></div>
        <div class="thumb"><a href="https://xcum.com/v/2/"><img alt="C" data-original="https://img/2.jpg">
          <span class="inf"><span>Other</span><span>10:00</span></span></a></div>
      </div>`;
    const items = parseVideoCards(site("xcum"), parse(html));
    expect(items[0].title).toBe("Real title");
    expect(items[0].duration_seconds).toBe(359);
    expect(items[1].duration_seconds).toBe(600);
  });

  test("model avatars inside a card don't become the poster or title", () => {
    const html = `
      <div class="thumbs">
        <div class="card"><a href="/video/1/x/"><img alt="Video one" src="https://cdn/v1.jpg"></a>
          <div class="card-meta"><a class="title" href="/video/1/x/">Video one</a>
          <a href="/models/asa/"><img alt="Asa" src="https://cdn/asa.jpg"> Asa</a></div></div>
        <div class="card"><a href="/video/2/y/"><img alt="Video two" src="https://cdn/v2.jpg"></a></div>
      </div>`;
    const items = parseVideoCards(site("w1mp"), parse(html));
    expect(items[0]).toMatchObject({ title: "Video one", poster_path: "https://cdn/v1.jpg" });
  });
});

describe("catalog entries", () => {
  test("index entries skip sort links and read counts", () => {
    const html = `
      <div class="headline"><ul class="sorting"><li><a href="/models/most-viewed/">Most Viewed</a></li></ul></div>
      <div class="list-models">
        <a href="/models/alice-frost/"><img src="/contents/models/1/s1.jpg" alt="Alice Frost"><span class="count">786</span></a>
        <a href="/models/bob/">Bob (12)</a>
        <a href="/models/2/">2</a>
      </div>`;
    const entries = parseCatalogEntries(MODELS, site("porngo"), parse(html));
    expect(entries).toEqual([
      {
        title: "Alice Frost",
        media_type: "performer",
        source_url: "https://www.porngo.com/models/alice-frost/",
        external_id: "kvs-porngo-model-alice-frost",
        poster_path: "https://www.porngo.com/contents/models/1/s1.jpg",
        count: 786,
      },
      {
        title: "Bob",
        media_type: "performer",
        source_url: "https://www.porngo.com/models/bob/",
        external_id: "kvs-porngo-model-bob",
        count: 12,
      },
    ]);
  });

  test("tag entries use the site's tag prefixes", () => {
    const html = `<div class="list-categories-sort"><a href="https://anysex.com/videos/categories/beauty/"><img alt="Beauty" src="/c/10.jpg"></a></div>`;
    const entries = parseCatalogEntries(TAGS, site("anysex"), parse(html));
    expect(entries.map((entry) => entry.external_id)).toEqual(["kvs-anysex-tag-beauty"]);
  });

  test("parseMinuteBadge reads a bare minute count only", () => {
    expect(parseMinuteBadge("3 min")).toBe(180);
    expect(parseMinuteBadge("45 minutes")).toBe(2700);
    expect(parseMinuteBadge("3 min ago")).toBe(0);
  });

  test("splitCount separates trailing counts", () => {
    expect(splitCount("1080p (896)")).toEqual(["1080p", 896]);
    expect(splitCount("Big Tits (1,234)")).toEqual(["Big Tits", 1234]);
    expect(splitCount("Plain")).toEqual(["Plain", 0]);
  });
});

describe("helpers", () => {
  test("interleave alternates sites", () => {
    const item = (id: string): DiscoveredItem => ({ title: id, media_type: "scene", source_url: id, external_id: id });
    const merged = interleave([[item("a1"), item("a2"), item("a3")], [item("b1")], [item("c1"), item("c2")]]);
    expect(merged.map((entry) => entry.external_id)).toEqual(["a1", "b1", "c1", "a2", "c2", "a3"]);
  });

  test("requestedQuality reads the download fragment", () => {
    expect(requestedQuality("https://ok.xxx/video/1/#720p")).toBe("720p");
    expect(requestedQuality("https://x.com/video/1/#High%20Quality")).toBe("High Quality");
    expect(requestedQuality("https://ok.xxx/video/1/")).toBe("");
  });

  test("mimeTypeOf recognises HLS by content type or URL", () => {
    expect(mimeTypeOf("https://cdn/x.mp4", "application/vnd.apple.mpegurl")).toBe("application/x-mpegURL");
    expect(mimeTypeOf("https://cdn.privatehost.com/hls/contents/videos/1/1,_360p.mp4,?x=1", "")).toBe(
      "application/x-mpegURL"
    );
    expect(mimeTypeOf("https://cdn/x.mp4?secure=1", "video/mp4")).toBe("video/mp4");
  });
});

describe("xgroovy", () => {
  test("cards read the video id, the src poster and the minute runtime", () => {
    const html = `
      <div class="list-videos"><div id="list_videos_custom_videos_list_items">
        <div class="item" data-video-id="795901">
          <a href="https://xgroovy.com/videos/795901/curvy-babe-shakes/" class="popito">
            <div class="img"><img class="thumb " src="https://i.xgroovy.com/contents/videos_screenshots/795000/795901/608x342/1.jpg" alt="Curvy Babe Shakes"></div>
            <strong class="title">Curvy Babe Shakes</strong>
          </a>
          <div class="wrap"><div class="duration">3 min</div></div>
        </div>
      </div></div>`;
    const items = parseListingCards(site("xgroovy"), parse(html));
    expect(items.map((item) => item.external_id)).toEqual(["kvs-xgroovy-795901"]);
    expect(items[0]!.title).toBe("Curvy Babe Shakes");
    expect(items[0]!.poster_path).toContain("795901/608x342/1.jpg");
    expect(items[0]!.duration_seconds).toBe(180);
  });

  test("video and listing URLs follow the /videos/<id>/ and /new/ scheme", () => {
    expect(videoKey(site("xgroovy"), "https://xgroovy.com/videos/770762/leather-clothed/")).toBe("770762");
    expect(site("xgroovy").latestPath).toBe("/new/");
    expect(siteForUrl("https://xgroovy.com/videos/1/x/")?.key).toBe("xgroovy");
  });
});

describe("javbangers members-only videos", () => {
  test("a private upload's page has no player, so it yields no sources", () => {
    const html = `
      <div class="no-player" style="width: 100%; height: 0; padding-bottom: 56%">
        <img src="https://jav.cdntrex.com/contents/videos_screenshots/391000/391800/preview.jpg" alt="Blonde">
        <span class="message">This video is a private video uploaded by FossilMan. Only active members can watch private videos.</span>
      </div>`;
    expect(extractSources(site("javbangers"), parse(html), html)).toEqual([]);
    expect(site("javbangers").membersOnlyVideos).toBe(true);
  });
});

describe("pimpbunny", () => {
  test("video pages are /videos/<slug>/, while /videos/<page>/ is a listing", () => {
    const pimpbunny = site("pimpbunny");
    expect(videoKey(pimpbunny, "https://pimpbunny.com/videos/alannasworldx-tittyfucks-her-boyfriend-s-cock/")).toBe(
      "alannasworldx-tittyfucks-her-boyfriend-s-cock"
    );
    expect(videoKey(pimpbunny, "https://pimpbunny.com/videos/2/")).toBe("");
    expect(videoKey(pimpbunny, "https://pimpbunny.com/videos/")).toBe("");
    expect(pimpbunny.latestPath).toBe("/videos/");
    expect(pimpbunny.searchPath).toBe("/search/{query}/");
  });

  test("related videos come from the similar-videos block", () => {
    const html = `
      <div id="list_videos_similar_videos"><div id="list_videos_similar_videos_items">
        <a href="https://pimpbunny.com/videos/angela-5-5-s-dildo-ride-under-the-sun/">Angela</a>
      </div></div>`;
    const blocks = parse(html).querySelectorAll(site("pimpbunny").relatedSelector);
    expect(blocks.length).toBeGreaterThan(0);
    expect(parseVideoCards(site("pimpbunny"), blocks[0]!).map((item) => item.external_id)).toEqual([
      "kvs-pimpbunny-angela-5-5-s-dildo-ride-under-the-sun",
    ]);
  });
});

describe("cam show archives", () => {
  test("cloudbate cards carry the cam username and the minute runtime", () => {
    const html = `
      <div class="row no-gutters" id="list_videos_latest_videos_list_items">
        <div class="item col-6"><div class="video-block">
          <span class="duration">12:58</span>
          <a class="thumb" href="https://www.cloudbate.com/video/1141719/elle-uwu-porn-cam-video-10-28-2026/" title="Elle_uwu porn cam video - 10/28/2026">
            <img class="video-img" src="https://www.cloudbate.com/contents/videos_screenshots/1141000/1141719/336x189/3.jpg" alt="elle_uwu recorded nude video 1141719">
          </a>
          <span class="streamer"><a href="https://www.cloudbate.com/models/elle_uwu/">elle_uwu</a></span>
        </div></div>
      </div>`;
    const items = parseListingCards(site("cloudbate"), parse(html));
    expect(items).toHaveLength(1);
    expect(items[0]!.external_id).toBe("kvs-cloudbate-1141719");
    expect(items[0]!.title).toBe("Elle_uwu porn cam video - 10/28/2026");
    expect(items[0]!.duration_seconds).toBe(12 * 60 + 58);
    expect(videoKey(site("cloudbate"), "https://www.cloudbate.com/models/elle_uwu/")).toBe("");
  });

  test("camstreams cards are /videos/<id>/ links with a lazy-loaded poster", () => {
    const html = `
      <div id="list_videos_latest_videos_list_items">
        <div class="item"><a href="https://camstreams.tv/videos/44484/lili-jones-mfc/" title="Lili jones mfc">
          <div class="img"><img class="thumb lazy-load" src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7" data-original="https://camstreams.tv/contents/videos_screenshots/44000/44484/320x180/1.jpg" alt="Lili jones mfc"></div>
          <strong class="title">Lili jones mfc</strong>
        </a></div>
      </div>`;
    const items = parseListingCards(site("camstreams"), parse(html));
    expect(items.map((item) => item.external_id)).toEqual(["kvs-camstreams-44484"]);
    expect(items[0]!.poster_path).toBe("https://camstreams.tv/contents/videos_screenshots/44000/44484/320x180/1.jpg");
    expect(site("camstreams").modelIndexPath).toBeUndefined();
  });

  test("a category naming a cam site credits it as the studio", () => {
    expect(camPlatformStudio(site("camstreams"), ["MyFreeCams", "fingering"])).toEqual({ name: "MyFreeCams" });
    expect(camPlatformStudio(site("cloudbate"), ["girls", "chaturbate"])).toEqual({ name: "Chaturbate" });
    expect(camPlatformStudio(site("camstreams"), ["fingering"])).toBeUndefined();
    expect(camPlatformStudio(site("okxxx"), ["Chaturbate"])).toBeUndefined();
  });
});
