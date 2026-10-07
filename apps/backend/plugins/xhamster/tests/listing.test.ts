import { test, expect } from "bun:test";
import { parse } from "node-html-parser";
import { parseVideoCards, thumbsFromInitials, landingStudio } from "../src/listing.ts";
import { parseInitials } from "../src/initials.ts";

// Two cards as xhamster serves them: the first fully rendered, the second a
// skeleton with no image, whose poster/runtime/channel only exist in the state.
const page = `<html><body>
<a class="video-thumb__image-container" href="https://xhamster.com/videos/first-video-xhAAA" aria-label="First video">
  <img class="thumb-image-container__image" src="https://ic-vt-nss.xhcdn.com/first.webp">
  <div data-role="video-duration">1:05</div>
</a>
<a class="video-thumb__image-container" href="https://xhamster.com/videos/second-video-xhBBB" aria-label="Second video">
  <div class="placeholderSkeleton"></div>
</a>
<script>window.initials={"searchResult":{"videoThumbProps":[
  {"pageURL":"https://xhamster.com/videos/first-video-xhAAA","thumbURL":"https://ic-vt-nss.xhcdn.com/first-state.webp","duration":65,"created":1629387339,
   "landing":{"type":"channel","name":"Xander-Vision","logo":"https://ic.xhcdn.com/logo.jpg","link":"https://xhamster.com/channels/xander-vision"}},
  {"pageURL":"https://xhamster.com/videos/second-video-xhBBB","thumbURL":"https://ic-vt-nss.xhcdn.com/second.webp","duration":753,"created":1629387339,
   "landing":{"type":"user","name":"SarberCorvinus","logo":null,"link":"https://xhamster.com/users/profiles/sarbercorvinus/videos"}}
]}};</script>
</body></html>`;

test("skeleton cards get poster, runtime, date and channel from the page state", () => {
  const cards = parseVideoCards(parse(page), thumbsFromInitials(parseInitials(page)));

  expect(cards.map((card) => card.external_id)).toEqual(["xhamster-xhAAA", "xhamster-xhBBB"]);
  expect(cards[0].poster_path).toBe("https://ic-vt-nss.xhcdn.com/first.webp");
  expect(cards[0].studio).toEqual({
    name: "Xander-Vision",
    source_url: "https://xhamster.com/channels/xander-vision",
    image_path: "https://ic.xhcdn.com/logo.jpg",
  });
  expect(cards[1]).toMatchObject({
    poster_path: "https://ic-vt-nss.xhcdn.com/second.webp",
    duration_seconds: 753,
    date: "2021-08-19",
    studio: { name: "SarberCorvinus", source_url: "https://xhamster.com/users/sarbercorvinus" },
  });
});

test("cards parse without page state, as before", () => {
  const cards = parseVideoCards(parse(page));
  expect(cards[1].poster_path).toBeUndefined();
  expect(cards[1].studio).toBeUndefined();
});

test("uploader listing links map to the profile URL the video page uses", () => {
  expect(
    landingStudio({ name: "SiennaDiamond", link: "https://xhamster.com/users/profiles/siennadiamond/videos" })
  ).toEqual({ name: "SiennaDiamond", source_url: "https://xhamster.com/users/siennadiamond" });
  expect(landingStudio({ name: "", link: "https://xhamster.com/channels/x" })).toBeUndefined();
});
