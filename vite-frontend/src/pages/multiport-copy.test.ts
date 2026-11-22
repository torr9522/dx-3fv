import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

test("multiport user copy consistently names the landing endpoint in Chinese", () => {
  const source = readFileSync(new URL("./multiport.tsx", import.meta.url), "utf8");

  for (const oldCopy of [
    "请填写 Landing IP 或域名",
    "Landing 仅支持 IPv4 地址或具有 A 记录的域名",
    "Landing 端口必须是 1-65535",
    "独立 Landing 业务",
    'label="Landing IP / 域名"',
    'label="Landing 端口"',
    "仅在 Landing 是受控",
  ]) {
    assert.doesNotMatch(source, new RegExp(oldCopy));
  }

  assert.match(source, /请填写落地 IP 或域名/);
  assert.match(source, /落地仅支持 IPv4 地址或具有 A 记录的域名/);
  assert.match(source, /label="落地 IP \/ 域名"/);
  assert.match(source, /label="落地端口"/);
});
