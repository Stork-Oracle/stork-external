# Sui gRPC v2 generated client code

The `v2/` directory contains vendored Go client code for the Sui fullnode gRPC API
(`sui.rpc.v2`), generated from Mysten's official proto definitions at
[MystenLabs/sui-apis](https://github.com/MystenLabs/sui-apis).

Last generated from sui-apis commit `baa7b24d643c9c90b371f645ef4c2bbf99af1775` (2026-09-04).

## Why gRPC

Sui deprecated JSON-RPC in 2026 (disabled on Sui Foundation fullnodes the week of
2026-07-27, code removal from `sui-node` mid-October 2026), with gRPC and GraphQL as
the replacements. See <https://docs.sui.io/develop/accessing-data/json-rpc-migration>.
The Sui chain pusher bindings (`../bindings`) use these generated clients for all
network calls; transaction building and signing remain on `go-sui-sdk`.

## Why vendored

Mysten does not publish generated Go SDKs (the buf.build module
`buf.build/mystenlabs/sui-apis` exists but has no generated Go SDK enabled), so the
`.pb.go` files are generated locally and committed. Only generated Go files are
vendored — no `.proto` files.

## Regenerating

```bash
./scripts/generate_sui_grpc.sh          # from main
./scripts/generate_sui_grpc.sh <commit> # pinned to a sui-apis commit
```

The script clones sui-apis, generates with buf + protoc-gen-go/protoc-gen-go-grpc
(versions pinned in the script; keep them compatible with the `go` directive in
go.mod), copies the output here, and rebuilds. Update the commit hash above after
regenerating.

One upstream quirk the script handles: sui-apis vendors copies of the protobuf
well-known types, and `google/protobuf/timestamp.proto` is the only one whose
`go_package` option still points at the deprecated `github.com/golang/protobuf`
module. The script rewrites it to `google.golang.org/protobuf/types/known/timestamppb`
before generating, and fails if any deprecated import survives.

## Lint

Generated files are excluded from linting via golangci-lint's `generated: lax`
detection, and `.golangci.yml` excludes the `rpc/v2` types from `exhaustruct` so
calling code can construct request messages without setting every field.
