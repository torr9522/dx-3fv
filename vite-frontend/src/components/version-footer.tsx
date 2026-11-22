import { useEffect, useState } from "react";

import { siteConfig } from "@/config/site";
import {
  AUTOMATIC_UPDATE_CHECK_CHANGED_EVENT,
  UPDATE_CHANNEL_CHANGED_EVENT,
  type UpdateReleaseChannel,
  getAutomaticUpdateCheckEnabled,
  getLatestVersionByChannel,
  getUpdateReleaseChannel,
  hasVersionUpdate,
} from "@/utils/version-update";

interface VersionFooterProps {
  version: string;
  containerClassName?: string;
  versionClassName?: string;
  poweredClassName?: string;
  updateBadgeClassName?: string;
}

export function VersionFooter({
  version,
  containerClassName,
  versionClassName,
  poweredClassName,
  updateBadgeClassName,
}: VersionFooterProps) {
  const [channel, setChannel] = useState<UpdateReleaseChannel>(
    getUpdateReleaseChannel(),
  );
  const [automaticCheckEnabled, setAutomaticCheckEnabled] = useState(
    getAutomaticUpdateCheckEnabled(),
  );
  const [updateAvailable, setUpdateAvailable] = useState(false);
  const [latestUpdateVersion, setLatestUpdateVersion] = useState<string | null>(
    null,
  );

  useEffect(() => {
    const handleChannelChange = () => {
      setChannel(getUpdateReleaseChannel());
      setAutomaticCheckEnabled(getAutomaticUpdateCheckEnabled());
    };

    window.addEventListener(UPDATE_CHANNEL_CHANGED_EVENT, handleChannelChange);
    window.addEventListener(
      AUTOMATIC_UPDATE_CHECK_CHANGED_EVENT,
      handleChannelChange,
    );
    window.addEventListener("storage", handleChannelChange);

    return () => {
      window.removeEventListener(
        UPDATE_CHANNEL_CHANGED_EVENT,
        handleChannelChange,
      );
      window.removeEventListener(
        AUTOMATIC_UPDATE_CHECK_CHANGED_EVENT,
        handleChannelChange,
      );
      window.removeEventListener("storage", handleChannelChange);
    };
  }, []);

  useEffect(() => {
    let active = true;

    if (!automaticCheckEnabled) {
      setUpdateAvailable(false);
      setLatestUpdateVersion(null);

      return () => {
        active = false;
      };
    }

    const checkUpdate = async () => {
      const latestVersion = await getLatestVersionByChannel(channel);

      if (!active) {
        return;
      }

      if (!latestVersion) {
        setUpdateAvailable(false);
        setLatestUpdateVersion(null);

        return;
      }

      const hasUpdate = hasVersionUpdate(version, latestVersion);

      setUpdateAvailable(hasUpdate);
      setLatestUpdateVersion(hasUpdate ? latestVersion : null);
    };

    void checkUpdate();

    return () => {
      active = false;
    };
  }, [automaticCheckEnabled, channel, version]);

  return (
    <div className={containerClassName}>
      <p className={versionClassName}>
        v{version}
        {updateAvailable && latestUpdateVersion && (
          <span className={updateBadgeClassName} role="status">
            {latestUpdateVersion}
          </span>
        )}
      </p>
      {siteConfig.hide_footer_brand !== true && (
        <p className={poweredClassName}>
          Powered by{" "}
          <a
            className="text-gray-500 dark:text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            href={siteConfig.github_repo}
            rel="noopener noreferrer"
            target="_blank"
          >
            DX-3FV
          </a>
        </p>
      )}
    </div>
  );
}
