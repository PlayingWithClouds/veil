import { test, expect, describe, afterEach } from "bun:test";
import { pluginSites, sitesFor } from "../src/sites.ts";

describe("sites picked by a search filter", () => {
  afterEach(() => {
    delete process.env.KVS_SITES;
  });

  test("meta lists every site by host, with its favicon", () => {
    const cloudbate = pluginSites().find((site) => site.host === "cloudbate.com");
    expect(cloudbate?.name).toBe("Cloudbate");
    expect(cloudbate?.icon).toBe("https://www.cloudbate.com/favicon.ico");
    expect(cloudbate?.content).toBe("cam-shows");
  });

  test("picked hosts are the only sites asked, even ones KVS_SITES leaves out", () => {
    process.env.KVS_SITES = "okxxx";
    const keys = sitesFor({ sites: ["cloudbate.com", "www.camstreams.tv"] }).map((site) => site.key);
    expect(keys.sort()).toEqual(["camstreams", "cloudbate"]);
  });

  test("without picks, the enabled sites are asked", () => {
    process.env.KVS_SITES = "okxxx, cloudbate.com";
    expect(sitesFor({}).map((site) => site.key).sort()).toEqual(["cloudbate", "okxxx"]);
  });

  test("picked sites still narrow to the content asked for", () => {
    expect(sitesFor({ sites: ["cloudbate.com", "ok.xxx"], content: "cam-shows" }).map((site) => site.key)).toEqual(["cloudbate"]);
  });
});
