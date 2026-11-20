# DX3FV-0004: Multiport management UI

- Priority: P1 feature
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.6-fix-active-path-identity`
- Release tag: `v0.1.7-feature-multiport-ui`

## Requirement

Expose the independent multiport business model without changing the original
FLVX Tunnel or Forward pages. Administrators must be able to find the feature,
configure a static GZ-to-IX path, and operate route lifecycle through the UI.

## Implementation

- Add the protected `/multiport` route.
- Add a desktop sidebar item without removing or renaming original entries.
- Add a mobile administrator entry under Profile.
- Group businesses by PathProfile.
- Keep Landing host and Landing port as separate inputs.
- Support empty/automatic and explicit/manual GZ and IX ports.
- Expose TCP, UDP, and TCP+UDP modes.
- Expose pause, resume, delete, and manual reconciliation actions.
- Keep IX NAT resource administration on the independent `/ix-ports` page.

The form presents GZ2 as a full-port policy instead of a fictitious narrow
range. Backend allocation still excludes SSH 22 and checks persisted and
observed listener ownership. Frontsea IX2 continues to use its separate NAT
resource declaration.

## Verification

```bash
cd vite-frontend
pnpm exec eslint --fix src/pages/multiport.tsx src/App.tsx \
  src/api/index.ts src/layouts/admin.tsx src/pages/profile.tsx
pnpm run build
```

Real-browser desktop/mobile CRUD acceptance is required after deployment. Safe
route editing, authoritative cumulative traffic, and segmented diagnostics are
intentionally separate features and are not claimed by this release.

## Rollback

Revert the commit pointed to by `v0.1.7-feature-multiport-ui`. The backend API
and persisted routes remain intact; only the new frontend entry and API client
bindings are removed.
