import type { Context } from "@neoworks/extension-system";
import { veilUrlOf } from "./config";
import type { VeilConfig } from "./config";
import { createVeilProvider } from "./provider";
import { createVeilApi } from "./veilApi";

export default {
  name: "veil",
  inject: ["sources", "http"],
  apply(ctx: Context, config: VeilConfig = {}) {
    const baseUrl = veilUrlOf(config);
    const provider = createVeilProvider(createVeilApi(baseUrl));
    ctx.effect(() => ctx.sources.register(provider), "source:veil");
    // The web picker builds thumbnail URLs against the veil server directly.
    ctx.effect(() => ctx.http.route("GET", "/api/plugins/veil/config", () => ({ url: baseUrl })), "route:veil-config");
  },
};
