import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const readPage = (name: string): string =>
  readFileSync(new URL(name, import.meta.url), "utf8");

test("all forwarding forms describe the shared automatic entry port policy", () => {
  const single = readPage("./single-forward.tsx");
  const tunnel = readPage("./forward.tsx");
  const multiport = readPage("./multiport.tsx");
  const config = readPage("./config.tsx");

  for (const source of [single, tunnel, multiport]) {
    assert.match(source, /随机分配 10000-65535 可用端口/);
  }
  assert.doesNotMatch(config, /顺次分配入口端口/);
  assert.doesNotMatch(tunnel, /留空则自动分配可用端口/);
});
