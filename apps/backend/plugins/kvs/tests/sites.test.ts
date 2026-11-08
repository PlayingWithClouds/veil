import { test, expect, describe } from "bun:test";
import { pluginSites, sitesFor } from "../src/sites.ts";

describe("sites picked by a search filter", () => {
  test("meta lists every site by host, scatporno off by default", () => {
    const sites = pluginSites();
    const cloudbate = sites.find((site) => site.host === "cloudbate.com");
    expect(cloudbate?.name).toBe("Cloudbate");
    expect(cloudbate?.icon).toBe("https://www.cloudbate.com/favicon.ico");
    expect(cloudbate?.off_by_default).toBe(false);
    expect(sites.find((site) => site.host === "scatporno.org")?.off_by_default).toBe(true);
  });

  test("picked hosts are the only sites asked, even one off by default", () => {
    const keys = sitesFor({ sites: ["cloudbate.com", "www.scatporno.org"] }).map((site) => site.key);
    expect(keys.sort()).toEqual(["cloudbate", "scatporno"]);
  });

  test("without picks, a site off by default isn't asked", () => {
    const keys = sitesFor({}).map((site) => site.key);
    expect(keys).toContain("cloudbate");
    expect(keys).not.toContain("scatporno");
  });

  test("picked sites still narrow to the content asked for", () => {
    expect(sitesFor({ sites: ["cloudbate.com", "scatporno.org"], content: "cam-shows" }).map((site) => site.key)).toEqual(["cloudbate"]);
  });
});
