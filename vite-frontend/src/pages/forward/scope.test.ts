import assert from "node:assert/strict";
import test from "node:test";

import { isTunnelForward } from "./scope.ts";

test("only positive Tunnel IDs belong on the Tunnel Forward page", () => {
  assert.equal(isTunnelForward(7), true);
  assert.equal(isTunnelForward(0), false);
  assert.equal(isTunnelForward(null), false);
  assert.equal(isTunnelForward(undefined), false);
});
