import assert from "node:assert/strict";
import test from "node:test";

import {
  isForwardRoute,
  readExpandedPreference,
  writeExpandedPreference,
} from "./forward-navigation.ts";

test("forward routes include the hub and all three child pages", () => {
  for (const path of [
    "/forwards",
    "/single-forward",
    "/forward",
    "/multiport",
  ]) {
    assert.equal(isForwardRoute(path), true, path);
  }
  assert.equal(isForwardRoute("/tunnel"), false);
});

test("an explicit collapse preference wins over active-route defaults", () => {
  const values = new Map<string, string>([["forward", "false"]]);
  const storage = {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
  };

  assert.equal(readExpandedPreference(storage, "forward", true), false);
  writeExpandedPreference(storage, "forward", true);
  assert.equal(readExpandedPreference(storage, "forward", false), true);
});

test("missing or invalid preferences use the caller default", () => {
  const storage = { getItem: () => "invalid" };

  assert.equal(readExpandedPreference(storage, "forward", true), true);
  assert.equal(readExpandedPreference(storage, "forward", false), false);
});
