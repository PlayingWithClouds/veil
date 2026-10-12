import { test, expect, describe } from "bun:test";
import { suggestionsFromHtml } from "../src/suggest.ts";

// A trimmed copy of what /suggest/isla/all/?layout=sections returns.
const SUGGESTIONS_HTML = `
<span id="qstit">Search suggestions:</span>
<div id="search-suggestions" class="qs-results" role="listbox">
<ul class="qs-keywords" role="presentation">
<li id="sugg1"><a href="/tag/isla-wild/" title="isla wild" id="sugga1"><span class="qslabel"><b class="qsmatch">isla</b> wild</span></a></li>
<li id="sugg2"><a href="/tag/island/" title="island" id="sugga2"><span class="qslabel"><b class="qsmatch">isla</b>nd</span></a></li>
</ul>
<ul class="qs-performers" role="presentation">
<li id="perf1"><a href="/pornstar/isla-summer/" title="Isla Summer" id="perfa1"><span class="qslabel">Isla Summer</span></a></li>
<li id="perf2"><a href="/pornstar/isla-summer/" title="isla summer" id="perfa2"></a></li>
</ul>
</div>`;

describe("suggestionsFromHtml", () => {
  test("keyword rows are searches and pornstar rows are performers", () => {
    expect(suggestionsFromHtml(SUGGESTIONS_HTML)).toEqual([
      { text: "isla wild", kind: "query" },
      { text: "island", kind: "query" },
      { text: "Isla Summer", kind: "performer" },
    ]);
  });

  test("a page without suggestions has none", () => {
    expect(suggestionsFromHtml("<div>nothing</div>")).toEqual([]);
  });
});
