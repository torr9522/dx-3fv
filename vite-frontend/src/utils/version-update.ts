export type UpdateReleaseChannel = "stable" | "dev";

export const UPDATE_CHANNEL_STORAGE_KEY = "update-release-channel";
export const UPDATE_CHANNEL_CHANGED_EVENT = "updateReleaseChannelChanged";
export const AUTOMATIC_UPDATE_CHECK_STORAGE_KEY =
  "dx3fv-automatic-update-check";
export const AUTOMATIC_UPDATE_CHECK_CHANGED_EVENT =
  "dx3fvAutomaticUpdateCheckChanged";
export const DX3FV_GITHUB_REPOSITORY = "torr9522/dx-3fv";

const CHANNEL_STABLE: UpdateReleaseChannel = "stable";
const CHANNEL_DEV: UpdateReleaseChannel = "dev";

const stableVersionPattern = /^\d+(?:\.\d+)+$/;
const testKeywordPattern = /(alpha|beta|rc)/i;

const VERSION_CACHE_TTL_MS = 10 * 60 * 1000;

type ReleaseItem = {
  tag_name?: string;
  draft?: boolean;
};

type LatestVersionCacheEntry = {
  value: string | null;
  expiresAt: number;
};

const latestVersionCache: Record<
  UpdateReleaseChannel,
  LatestVersionCacheEntry
> = {
  stable: { value: null, expiresAt: 0 },
  dev: { value: null, expiresAt: 0 },
};

const normalizeChannel = (
  value: string | null | undefined,
): UpdateReleaseChannel => {
  return value === CHANNEL_DEV ? CHANNEL_DEV : CHANNEL_STABLE;
};

export const getUpdateReleaseChannel = (): UpdateReleaseChannel => {
  if (typeof window === "undefined") {
    return CHANNEL_STABLE;
  }

  return normalizeChannel(localStorage.getItem(UPDATE_CHANNEL_STORAGE_KEY));
};

export const setUpdateReleaseChannel = (
  channel: UpdateReleaseChannel,
): void => {
  if (typeof window === "undefined") {
    return;
  }

  localStorage.setItem(UPDATE_CHANNEL_STORAGE_KEY, normalizeChannel(channel));
  window.dispatchEvent(new Event(UPDATE_CHANNEL_CHANGED_EVENT));
};

export const getAutomaticUpdateCheckEnabled = (): boolean => {
  if (typeof window === "undefined") return false;

  return localStorage.getItem(AUTOMATIC_UPDATE_CHECK_STORAGE_KEY) === "true";
};

export const setAutomaticUpdateCheckEnabled = (enabled: boolean): void => {
  if (typeof window === "undefined") return;

  localStorage.setItem(
    AUTOMATIC_UPDATE_CHECK_STORAGE_KEY,
    enabled ? "true" : "false",
  );
  window.dispatchEvent(new Event(AUTOMATIC_UPDATE_CHECK_CHANGED_EVENT));
};

const normalizeTag = (tag: string): string => {
  return tag.trim().replace(/^v/i, "");
};

type ReleaseTagChannel = UpdateReleaseChannel | null;

const releaseChannelFromTag = (tag: string): ReleaseTagChannel => {
  const normalizedTag = normalizeTag(tag).toLowerCase();

  if (!normalizedTag) {
    return null;
  }

  if (stableVersionPattern.test(normalizedTag)) {
    return CHANNEL_STABLE;
  }

  if (testKeywordPattern.test(normalizedTag)) {
    return CHANNEL_DEV;
  }

  return null;
};

type VersionParts = {
  numbers: number[];
  stageRank: number;
  stageNumber: number;
};

const parseVersionParts = (version: string): VersionParts => {
  const normalized = normalizeTag(version).toLowerCase();
  const numberMatches = normalized.match(/\d+/g) || [];
  const numbers = numberMatches.map((item) => Number.parseInt(item, 10));

  let stageRank = 0;

  if (normalized.includes("rc")) {
    stageRank = 3;
  } else if (normalized.includes("beta")) {
    stageRank = 2;
  } else if (normalized.includes("alpha")) {
    stageRank = 1;
  } else if (stableVersionPattern.test(normalized)) {
    stageRank = 4;
  }

  const stageNumberMatch = normalized.match(/(?:alpha|beta|rc)[.-]?(\d+)/);
  const stageNumber = stageNumberMatch
    ? Number.parseInt(stageNumberMatch[1], 10)
    : 0;

  return {
    numbers,
    stageRank,
    stageNumber,
  };
};

export const compareVersions = (left: string, right: string): number => {
  const a = parseVersionParts(left);
  const b = parseVersionParts(right);
  const maxLength = Math.max(a.numbers.length, b.numbers.length);

  for (let i = 0; i < maxLength; i += 1) {
    const aValue = a.numbers[i] || 0;
    const bValue = b.numbers[i] || 0;

    if (aValue !== bValue) {
      return aValue - bValue;
    }
  }

  if (a.stageRank !== b.stageRank) {
    return a.stageRank - b.stageRank;
  }

  if (a.stageNumber !== b.stageNumber) {
    return a.stageNumber - b.stageNumber;
  }

  return 0;
};

export const getLatestVersionByChannel = async (
  channel: UpdateReleaseChannel,
): Promise<string | null> => {
  const normalizedChannel = normalizeChannel(channel);
  const now = Date.now();
  const cached = latestVersionCache[normalizedChannel];

  if (cached.value && cached.expiresAt > now) {
    return cached.value;
  }

  const response = await fetch(
    `https://api.github.com/repos/${DX3FV_GITHUB_REPOSITORY}/releases?per_page=50`,
    {
      headers: {
        Accept: "application/vnd.github+json",
      },
    },
  );

  if (!response.ok) {
    return null;
  }

  const releases = (await response.json()) as ReleaseItem[];
  const candidateTags = releases
    .filter((release) => !release.draft && typeof release.tag_name === "string")
    .map((release) => (release.tag_name || "").trim())
    .filter((tag) => releaseChannelFromTag(tag) === normalizedChannel);

  if (candidateTags.length === 0) {
    return null;
  }

  const latest = candidateTags.sort((a, b) => compareVersions(b, a))[0];

  latestVersionCache[normalizedChannel] = {
    value: latest,
    expiresAt: now + VERSION_CACHE_TTL_MS,
  };

  return latest;
};

export const hasVersionUpdate = (
  currentVersion: string,
  latestVersion: string,
): boolean => {
  return (
    compareVersions(normalizeTag(currentVersion), normalizeTag(latestVersion)) <
    0
  );
};
