import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

import {
  FIXED_TUNNEL_STRATEGY,
  isLegacyMultiNodeTunnelPath,
  validateTunnelForm,
} from "./form.ts";

const onlineNodes = [
  { id: 1, status: 1 },
  { id: 2, status: 1 },
  { id: 3, status: 1 },
];

const baseTunnel = {
  name: "fixed-path",
  type: 2,
  inNodeId: [{ nodeId: 1 }],
  outNodeId: [{ nodeId: 2 }],
  chainNodes: [],
  trafficRatio: 1,
};

test("new tunnels use the fixed fifo selector and reject multi-node paths", () => {
  assert.equal(FIXED_TUNNEL_STRATEGY, "fifo");
  assert.equal(isLegacyMultiNodeTunnelPath(baseTunnel.outNodeId), false);

  const errors = validateTunnelForm(
    {
      ...baseTunnel,
      outNodeId: [{ nodeId: 2 }, { nodeId: 3 }],
      chainNodes: [[{ nodeId: 2 }, { nodeId: 3 }]],
    },
    onlineNodes,
  );

  assert.match(errors.outNodeId || "", /固定出口/);
  assert.match(errors.chainNodes || "", /固定节点/);
});

test("editing a legacy multi-node tunnel remains compatible", () => {
  const legacy = {
    ...baseTunnel,
    outNodeId: [{ nodeId: 2 }, { nodeId: 3 }],
  };

  assert.equal(isLegacyMultiNodeTunnelPath(legacy.outNodeId), true);
  assert.equal(
    validateTunnelForm(legacy, onlineNodes, true).outNodeId,
    undefined,
  );
});

test("Tunnel form no longer exposes load-balancing controls", () => {
  const source = readFileSync(
    new URL("../tunnel.tsx", import.meta.url),
    "utf8",
  );

  assert.doesNotMatch(source, /label="负载策略"/);
  assert.doesNotMatch(source, /<SelectItem key="fifo">主备/);
  assert.doesNotMatch(source, /<SelectItem key="round">轮询/);
  assert.doesNotMatch(source, /<SelectItem key="rand">随机/);
  assert.doesNotMatch(source, /<SelectItem key="best">最优/);
});
