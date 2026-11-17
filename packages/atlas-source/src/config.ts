export interface VeilConfig {
  /** Base URL of the veil server. */
  url?: string;
}

const DEFAULT_VEIL_URL = "http://localhost:8080";

export function veilUrlOf(config: VeilConfig): string {
  if (typeof config.url !== "string" || config.url.trim() === "") {
    return DEFAULT_VEIL_URL;
  }
  return config.url.trim().replace(/\/+$/, "");
}
