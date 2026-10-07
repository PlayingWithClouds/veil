import { test, expect, describe } from "bun:test";
import { creditedStudio, extractPerformers, type ApiVideoDetail } from "../src/scrape.ts";

/** The `video` record of a saved video API response. */
async function loadVideo(name: string): Promise<ApiVideoDetail> {
  const response = await Bun.file(`${import.meta.dir}/fixtures/${name}.json`).json();
  return response.video;
}

// Channel-published, `models` empty: performers come from models_suggested.
const channelVideo = await loadVideo("video-channel");
// Channel-published with linked models.
const channelModelsVideo = await loadVideo("video-channel-models");
// Member upload on txxx ("UsersUpload" source group, no channel).
const memberUploadVideo = await loadVideo("video-member-upload");
// Member upload on hclips ("UserUpload" source group, no channel).
const hclipsUploadVideo = await loadVideo("video-hclips-userupload");

describe("extractPerformers", () => {
  test("uses linked models with their /models/ page", () => {
    expect(extractPerformers("txxx.com", channelModelsVideo)).toEqual([
      { name: "Angelo Godshack", order: 0, source_url: "https://txxx.com/models/angelo-godshack/" },
      { name: "Tru Kait", order: 1, source_url: "https://txxx.com/models/tru-kait/" },
    ]);
  });

  test("falls back to models_suggested when models is empty", () => {
    expect(extractPerformers("txxx.com", channelVideo)).toEqual([
      { name: "Chanel Preston", order: 0 },
      { name: "Tasha Reign", order: 1 },
    ]);
    expect(extractPerformers("txxx.com", memberUploadVideo)).toEqual([{ name: "James Bang", order: 0 }]);
  });

  test("returns nothing when neither list names anyone", () => {
    expect(extractPerformers("hclips.com", hclipsUploadVideo)).toEqual([]);
  });
});

describe("creditedStudio", () => {
  test("credits the channel under the site", () => {
    expect(creditedStudio("txxx.com", channelVideo)).toEqual({
      type: "studio",
      external_id: "txxx-channel-250@txxx.com",
      source_url: "https://txxx.com/channel/brazzers/",
      name: "Brazzers Network",
      image_path: "https://tn.txxx.tube/contents/cst/250/c3_271064.jpg",
      parent: "Txxx",
    });
  });

  test("credits the uploading member when there is no channel or real source", () => {
    expect(creditedStudio("txxx.com", memberUploadVideo)).toEqual({
      type: "studio",
      external_id: "txxx-member-6098065@txxx.com",
      source_url: "https://txxx.com/members/6098065/",
      name: "SeeHimFuck",
      image_path: "https://tn.txxx.tube/contents/avatars/upd/1492700652/6098000/6098065.png",
      parent: "Txxx",
    });
    expect(creditedStudio("hclips.com", hclipsUploadVideo)).toMatchObject({
      source_url: "https://hclips.com/members/3100911/",
      name: "ina37",
      parent: "Hclips",
    });
  });

  test("credits a content source, then a real source group, before the member", () => {
    const withoutChannel = { ...channelVideo, channel: [] };

    expect(creditedStudio("txxx.com", { ...withoutChannel, source_title: "BBC Pie", source_dir: "bbcpie-com" }))
      .toMatchObject({ name: "BBC Pie", source_url: "https://txxx.com/channel/bbcpie-com/", parent: "Txxx" });
    expect(creditedStudio("txxx.com", withoutChannel)).toMatchObject({
      external_id: "txxx-source-group-adultforce-com@txxx.com",
      source_url: "",
      name: "adultforce.com",
      parent: "Txxx",
    });
  });

  test("returns null when the API names no publisher", () => {
    const anonymous = { ...memberUploadVideo, user: null };
    expect(creditedStudio("txxx.com", anonymous)).toBeNull();
  });
});
