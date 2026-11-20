# Third-Party Notices

DX-3FV includes or derives from the following projects. This list supplements,
and does not replace, the full dependency inventories in `sbom/`.

## flux-panel / FLVX

DX-3FV derives from FLVX, which derives from flux-panel. The original
flux-panel attribution and Apache License 2.0 are preserved in `NOTICE` and
`LICENSE-APACHE`. DX-3FV and FLVX modifications are distributed under GPLv3;
see `LICENSE`.

## GOST / go-gost

The embedded forwarding runtime in `go-gost/` and `go-gost/x/` derives from
go-gost software, Copyright (c) 2016 ginuerzh, licensed under the MIT License.
See `LICENSES/GOST-MIT.txt`.

## Next UI bridge

Frontend bridge code includes material from Next UI, Copyright (c) 2024 Next
UI, licensed under the MIT License. See `LICENSES/NEXTUI-MIT.txt`.

## Package dependencies

Go and npm packages retain their respective licenses and notices. Exact module
versions and the sanitized frontend production-license inventory are generated
under `sbom/` in each binary release. Common license families include MIT,
Apache-2.0, ISC, BSD, MPL-2.0, 0BSD, and BlueOak-1.0.0.
