import { test, expect } from "bun:test";
import {
  parseInitials,
  extractPerformers,
  extractTags,
  extractStudio,
  extractUploader,
  extractDetails,
  extractDate,
} from "../src/initials.ts";

const channelFixture = await Bun.file(`${import.meta.dir}/fixtures/video-initials.html`).text();
const uploaderFixture = await Bun.file(`${import.meta.dir}/fixtures/video-uploader.html`).text();

test("parseInitials reads the embedded window.initials JSON", () => {
  const initials = parseInitials(channelFixture);
  expect(initials?.videoTagsComponent?.tags?.length).toBe(8);
  expect(parseInitials("<html><body>no state</body></html>")).toBeUndefined();
  expect(parseInitials("<script>window.initials={broken;</script>")).toBeUndefined();
});

test("performers come only from pornstar chips, not sidebar links", () => {
  expect(extractPerformers(parseInitials(channelFixture))).toEqual([
    { name: "Xander Corvus", order: 0, source_url: "https://xhamster.com/pornstars/xander-corvus" },
  ]);
  expect(extractPerformers(parseInitials(uploaderFixture))).toEqual([]);
});

test("tags are the category and tag chips, de-duplicated, without nav noise", () => {
  expect(extractTags(parseInitials(channelFixture))).toEqual(["In English", "4K Porn", "Blowjob", "Big Dick"]);
});

test("studio is the channel, never a brand or a member", () => {
  expect(extractStudio(parseInitials(channelFixture))).toEqual({
    name: "Xander-Vision",
    source_url: "https://xhamster.com/channels/xander-vision",
  });
  expect(extractStudio(parseInitials(uploaderFixture))).toBeUndefined();
  expect(extractStudio(undefined)).toBeUndefined();
});

test("uploader is the member who posted it, not a channel", () => {
  expect(extractUploader(parseInitials(uploaderFixture))).toEqual({
    name: "SiennaDiamond",
    url: "https://xhamster.com/users/siennadiamond",
  });
  expect(extractUploader(parseInitials(channelFixture))).toBeUndefined();
  expect(extractUploader(undefined)).toBeUndefined();
});

test("details and date come from the video model", () => {
  const initials = parseInitials(channelFixture);
  expect(extractDetails(initials)).toBe(
    "Horny Eva Nexus starts licking Xander's cock and gives a deep blowjob by getting on her knees!"
  );
  expect(extractDate(initials)).toBe("2026-09-28");
  expect(extractDetails(parseInitials(uploaderFixture))).toBeUndefined();
});
