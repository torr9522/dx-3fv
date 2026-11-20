import assert from "node:assert/strict";
import test from "node:test";

import { isPathCollapsed, togglePathCollapsed } from "./multiport-collapse.ts";

test("paths are expanded by default and collapse independently", () => {
  let state = {};
  assert.equal(isPathCollapsed(state, 10), false);
  state = togglePathCollapsed(state, 10);
  state = togglePathCollapsed(state, 20);
  assert.equal(isPathCollapsed(state, 10), true);
  assert.equal(isPathCollapsed(state, 20), true);
  state = togglePathCollapsed(state, 10);
  assert.equal(isPathCollapsed(state, 10), false);
  assert.equal(isPathCollapsed(state, 20), true);
});
