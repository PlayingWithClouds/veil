export const SITE = "https://chaturbate.com";

export const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

/** Headers the site's own pages send on their XHR calls; the endpoints answer a login redirect without them. */
export const API_HEADERS: Record<string, string> = {
  "User-Agent": UA,
  "X-Requested-With": "XMLHttpRequest",
  Accept: "application/json",
};
