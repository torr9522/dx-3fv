# DX3FV-0025: Remove the page background flash during theme switching

- Priority: P2 UI regression
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.28-white-panel-background`
- Release tag: `v0.1.29-fix-theme-switch-background-flash`
- Schema change: none

## Problem

Real Chromium acceptance found that switching from dark mode back to light
mode removed the root `dark` class immediately, but the document body remained
black and then animated through grey values for about 200 ms. The inline
boot-time stylesheet applied transitions to both the body colours and the root
`color-scheme`, contradicting its stated purpose of preventing theme flashes.

## Evidence and root cause

Frame sampling after the light-mode click observed non-white body backgrounds
after the root had already entered light mode: `rgb(0, 0, 0)`, intermediate
grey values, and finally `rgb(255, 255, 255)`. Computed styles identified:

- `body transition: background-color 0.2s ease, color 0.2s ease`
- `html transition: color-scheme 0.2s ease`

Both declarations came from `vite-frontend/index.html`.

## Fix

Remove only the document-level body colour and root colour-scheme transitions.
Theme class and token changes are now atomic. Component hover, modal, button,
and other interaction animations remain unchanged.

## Verification

- Frontend type checking and production build pass.
- Chromium frame sampling finds no non-white frame after light mode becomes
  active.
- Desktop and mobile route, modal, overflow, and screenshot acceptance pass.
- Dark mode continues to render its original dark background.

## Rollback

Revert the commit tagged `v0.1.29-fix-theme-switch-background-flash`. No
database or backend rollback is required.
