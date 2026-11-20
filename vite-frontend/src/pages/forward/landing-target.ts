export interface LandingTarget {
  host: string;
  port: string;
}

export interface LandingTargetValidation {
  fieldErrors: Array<{ host?: string; port?: string }>;
  valid: boolean;
}

export const emptyLandingTarget = (): LandingTarget => ({ host: "", port: "" });

const parseStoredTarget = (value: string): LandingTarget => {
  const target = value.trim();
  const bracketed = target.match(/^\[([^\]]+)]:(\d+)$/);
  if (bracketed) return { host: bracketed[1], port: bracketed[2] };

  const separator = target.lastIndexOf(":");
  if (separator <= 0) return { host: target, port: "" };
  return {
    host: target.slice(0, separator),
    port: target.slice(separator + 1),
  };
};

export const parseLandingTargets = (value: string): LandingTarget[] => {
  const targets = value
    .split(",")
    .map(parseStoredTarget)
    .filter((target) => target.host || target.port);
  return targets.length > 0 ? targets : [emptyLandingTarget()];
};

export const formatLandingTargets = (targets: LandingTarget[]): string =>
  targets
    .map(({ host, port }) => {
      const normalizedHost = host.trim();
      const storedHost = normalizedHost.includes(":")
        ? `[${normalizedHost.replace(/^\[|]$/g, "")}]`
        : normalizedHost;
      return `${storedHost}:${port.trim()}`;
    })
    .join(",");

const isValidIPv4 = (host: string): boolean => {
  const parts = host.split(".");
  return parts.length === 4 && parts.every((part) => /^\d{1,3}$/.test(part) && Number(part) <= 255);
};

const isValidDomain = (host: string): boolean => {
  if (host.length > 253 || !host.includes(".") || /^[\d.]+$/.test(host)) return false;
  return host.split(".").every((label) =>
    label.length >= 1 &&
    label.length <= 63 &&
    /^[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?$/.test(label),
  );
};

export const isValidIPv4OnlyLandingHost = (value: string): boolean => {
  const host = value.trim().replace(/^\[|]$/g, "");

  return isValidIPv4(host) || isValidDomain(host);
};

export const validateLandingTargets = (
  targets: LandingTarget[],
): LandingTargetValidation => {
  const seen = new Set<string>();
  const fieldErrors = targets.map((target) => {
    const errors: { host?: string; port?: string } = {};
    const host = target.host.trim().replace(/^\[|]$/g, "");
    const port = Number(target.port);

    if (!host) errors.host = "请输入落地 IP 或域名";
    else if (!isValidIPv4OnlyLandingHost(host)) {
      errors.host = "请输入有效的 IPv4 地址或具有 A 记录的域名";
    }

    if (!/^\d+$/.test(target.port.trim()) || port < 1 || port > 65535) {
      errors.port = "端口必须为 1-65535";
    }

    if (!errors.host && !errors.port) {
      const key = `${host.toLowerCase()}:${port}`;
      if (seen.has(key)) errors.host = "落地目标不能重复";
      seen.add(key);
    }
    return errors;
  });

  return {
    fieldErrors,
    valid: targets.length > 0 && fieldErrors.every((item) => !item.host && !item.port),
  };
};
