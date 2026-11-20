export const isTunnelForward = (tunnelId: number | null | undefined): boolean =>
  Number(tunnelId || 0) > 0;
