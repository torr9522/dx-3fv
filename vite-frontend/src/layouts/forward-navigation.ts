export const FORWARD_NAV_STORAGE_KEY = "dx3fv_forward_navigation_expanded";
export const MOBILE_FORWARD_NAV_STORAGE_KEY =
  "dx3fv_mobile_forward_navigation_expanded";

const FORWARD_ROUTES = new Set([
  "/forwards",
  "/single-forward",
  "/forward",
  "/multiport",
]);

export function isForwardRoute(pathname: string): boolean {
  return FORWARD_ROUTES.has(pathname);
}

export function readExpandedPreference(
  storage: Pick<Storage, "getItem">,
  key: string,
  defaultExpanded: boolean,
): boolean {
  const stored = storage.getItem(key);

  if (stored === "true") return true;
  if (stored === "false") return false;

  return defaultExpanded;
}

export function writeExpandedPreference(
  storage: Pick<Storage, "setItem">,
  key: string,
  expanded: boolean,
): void {
  storage.setItem(key, String(expanded));
}
