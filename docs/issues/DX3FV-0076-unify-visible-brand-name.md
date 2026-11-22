# DX3FV-0076: Unify the visible product brand name

- Priority: P2
- Affected versions: v1.0.0-v1.0.4
- State: FIXED_IN_DX3FV
- GitHub: https://github.com/torr9522/dx-3fv/issues/20

## Root cause

The DX-3FV publication history retained upstream FLVX display defaults in the
HTML shell, PWA manifest, runtime site configuration, footer, settings copy and
built-in theme metadata. Internal compatibility identifiers were intentionally
unchanged, but the visible fallback brand was inconsistent with the product.

## Fix

- Use `DX-3FV` for the browser title, application metadata and PWA name.
- Use `DX-3FV` as the non-commercial runtime brand fallback shown by login,
  desktop navigation and mobile navigation.
- Update footer, settings and built-in theme display copy.
- Keep the `flvx` import-format key but present it as a neutral compatibility
  format, avoiding an upstream brand in user-visible copy.
- Add a static display-brand contract covering the release frontend inputs.

## Preserved compatibility identifiers

This change does not rename `flvx:*` browser storage keys, `data-flvx-theme`,
the `flvx` import-format value, `FLVX-` license keys, environment variables,
API headers, Go modules, nftables resources, service names, paths, executables,
database identifiers, upstream documentation or license notices.

## Rollback

Redeploy v1.0.4. Schema 12, backend behavior, Agent and Runtime are unchanged.
