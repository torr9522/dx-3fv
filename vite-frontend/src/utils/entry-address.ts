import type { NodeApiItem } from "@/api/types";

const invalidHostnames = new Set(["localhost", "localhost.localdomain"]);

const normalizeIPv4 = (value: string): number[] | null => {
  const parts = value.split(".");

  if (parts.length !== 4) return null;
  const octets = parts.map((part) => Number(part));

  return octets.every(
    (octet, index) =>
      Number.isInteger(octet) &&
      octet >= 0 &&
      octet <= 255 &&
      String(octet) === parts[index],
  )
    ? octets
    : null;
};

const isNonPublicIPv4 = ([a, b]: number[]) =>
  a === 0 ||
  a === 10 ||
  a === 127 ||
  (a === 100 && b >= 64 && b <= 127) ||
  (a === 169 && b === 254) ||
  (a === 172 && b >= 16 && b <= 31) ||
  (a === 192 && b === 168) ||
  a >= 224;

export const normalizePublicEntryAddress = (value: unknown): string | null => {
  if (typeof value !== "string") return null;

  const host = value.trim().replace(/^\[|\]$/g, "");
  const lower = host.toLowerCase();

  if (
    !host ||
    invalidHostnames.has(lower) ||
    lower.endsWith(".local") ||
    host === "::" ||
    host === "::1" ||
    host.includes(":")
  ) {
    return null;
  }

  const ipv4 = normalizeIPv4(host);

  if (ipv4) return isNonPublicIPv4(ipv4) ? null : host;
  if (/^[0-9.]+$/.test(host)) return null;

  return /^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$/i.test(host) ? host : null;
};

export const getNodePublicEntryAddress = (
  node: Pick<NodeApiItem, "serverIp" | "serverIpV4"> | undefined,
): string | null => {
  return (
    normalizePublicEntryAddress(node?.serverIpV4) ||
    normalizePublicEntryAddress(node?.serverIp)
  );
};
