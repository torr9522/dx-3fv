# DX3FV-0024: Use a pure-white light panel background

- Priority: P3 visual requirement
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.27-restore-business-diagnosis`
- Release tag: `v0.1.28-white-panel-background`
- Schema change: none

## Problem

The stock light theme used `#f6f7fb` for the document background,
`#f2f2f7` for the main layout, and a full-screen blue-grey wallpaper pseudo
element. The combined result did not meet the DX-3FV pure-white background
requirement and produced different colours between the document, scroll area,
and layout container.

## Fix

- Set the base light `--background` token to `#FFFFFF`.
- Set the registered default theme's light background token to `#FFFFFF` so
  theme activation cannot restore the old grey-blue value.
- Make the main mesh layout consume the background token.
- Remove only the light-mode wallpaper layer.
- Preserve the existing dark theme colours and dark wallpaper.
- Preserve explicit administrator-selected custom colours and wallpapers.

No navigation, layout structure, component, route, or interaction was changed.

## Verification

- Production frontend type checking and build pass.
- Chromium checks computed styles and sampled pixels at 1440px and 390px.
- Dashboard, Forward, Tunnel, Node, Multiport, IX Ports, Monitor, Settings,
  modal, mobile navigation, scrolling, overflow, and browser errors are
  checked in the real test Panel.
- The independent FLVX reference Panel remains unchanged.

## Rollback

Revert the commit tagged `v0.1.28-white-panel-background`. This restores the
stock light wallpaper and grey-blue background without affecting backend data.
