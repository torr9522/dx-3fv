import assert from "node:assert/strict";
import test from "node:test";

import {
  formatLandingTargets,
  parseLandingTargets,
  validateLandingTargets,
} from "./landing-target.ts";

test("parses legacy IPv4, domain, and bracketed IPv6 targets", () => {
  assert.deepEqual(parseLandingTargets("192.0.2.1:443,example.com:8443,[2001:db8::10]:9443"), [
    { host: "192.0.2.1", port: "443" },
    { host: "example.com", port: "8443" },
    { host: "2001:db8::10", port: "9443" },
  ]);
});

test("serializes structured targets to the compatible backend format", () => {
  assert.equal(formatLandingTargets([
    { host: "192.0.2.1", port: "443" },
    { host: "2001:db8::10", port: "8443" },
  ]), "192.0.2.1:443,[2001:db8::10]:8443");
});

test("validates host, port, and duplicate targets independently", () => {
  const result = validateLandingTargets([
    { host: "example.com", port: "443" },
    { host: "EXAMPLE.com", port: "443" },
    { host: "bad host", port: "70000" },
  ]);
  assert.equal(result.valid, false);
  assert.equal(result.fieldErrors[1].host, "落地目标不能重复");
  assert.match(result.fieldErrors[2].host || "", /有效/);
  assert.match(result.fieldErrors[2].port || "", /1-65535/);
});

test("does not accept an invalid numeric address as a domain", () => {
  const result = validateLandingTargets([{ host: "999.1", port: "443" }]);
  assert.equal(result.valid, false);
  assert.match(result.fieldErrors[0].host || "", /有效/);
});

test("keeps legacy IPv6 parsing but rejects it for new IPv4-only targets", () => {
  const result = validateLandingTargets([{ host: "2001:db8::10", port: "443" }]);
  assert.equal(result.valid, false);
  assert.match(result.fieldErrors[0].host || "", /IPv4/);
});
