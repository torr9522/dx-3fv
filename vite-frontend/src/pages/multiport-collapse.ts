export type CollapsedPathMap = Record<string, boolean>;

export const isPathCollapsed = (
  state: CollapsedPathMap,
  pathId: number,
): boolean => state[String(pathId)] === true;

export const togglePathCollapsed = (
  state: CollapsedPathMap,
  pathId: number,
): CollapsedPathMap => {
  const key = String(pathId);
  if (state[key]) {
    const next = { ...state };
    delete next[key];
    return next;
  }
  return { ...state, [key]: true };
};
