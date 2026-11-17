import type {} from "@atlas/contracts/web";
import type { Context } from "@neoworks/extension-system";
import GalleryPicker from "./GalleryPicker.svelte";
import RandomPicker from "./RandomPicker.svelte";
import ScenePicker from "./ScenePicker.svelte";

export default {
  name: "veil",
  inject: ["sourcePickers"],
  apply(ctx: Context) {
    ctx.effect(
      () => ctx.sourcePickers.register({ id: "veil:gallery", sourceKind: "veil:gallery", component: GalleryPicker }),
      "picker:veil-gallery",
    );
    ctx.effect(
      () => ctx.sourcePickers.register({ id: "veil:scene", sourceKind: "veil:scene", component: ScenePicker }),
      "picker:veil-scene",
    );
    ctx.effect(
      () => ctx.sourcePickers.register({ id: "veil:random", sourceKind: "veil:random", component: RandomPicker }),
      "picker:veil-random",
    );
  },
};
