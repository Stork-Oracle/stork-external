#!/usr/bin/env bash
#
# Regenerates the vendored Sui gRPC v2 Go client code in
# apps/chain_pusher/pkg/sui/rpc/v2 from Mysten's official proto definitions
# (https://github.com/MystenLabs/sui-apis).
#
# See apps/chain_pusher/pkg/sui/rpc/README.md for background.
#
# Usage:
#   ./scripts/generate_sui_grpc.sh [sui-apis-ref]
#
# sui-apis-ref defaults to main. Pass a commit hash to pin the proto source.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${REPO_ROOT}/apps/chain_pusher/pkg/sui/rpc/v2"
SUI_APIS_REF="${1:-main}"

# Pinned tool versions. protoc-gen-go-grpc and buf are pinned to releases that
# build with the go directive in go.mod (do not bump past what go.mod's Go
# version supports).
PROTOC_GEN_GO_VERSION="v1.36.5"
PROTOC_GEN_GO_GRPC_VERSION="v1.5.1"
BUF_VERSION="v1.47.2"

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

echo "Installing protoc plugins..."
go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"
go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}"

echo "Cloning MystenLabs/sui-apis @ ${SUI_APIS_REF}..."
git clone --quiet https://github.com/MystenLabs/sui-apis.git "${WORK_DIR}/sui-apis"
git -C "${WORK_DIR}/sui-apis" checkout --quiet "${SUI_APIS_REF}"
echo "sui-apis commit: $(git -C "${WORK_DIR}/sui-apis" rev-parse HEAD)"

# Upstream quirk: sui-apis vendors copies of the protobuf well-known types, and
# timestamp.proto is the only one whose go_package still points at the
# deprecated github.com/golang/protobuf module. Rewrite it to the modern path
# so generated code does not import the deprecated shim.
sed -i.bak \
  's|option go_package = "github.com/golang/protobuf/ptypes/timestamp";|option go_package = "google.golang.org/protobuf/types/known/timestamppb";|' \
  "${WORK_DIR}/sui-apis/proto/google/protobuf/timestamp.proto"

cat > "${WORK_DIR}/sui-apis/buf.gen.yaml" <<'EOF'
version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg
    - file_option: go_package
      path: google/rpc/status.proto
      value: google.golang.org/genproto/googleapis/rpc/status
plugins:
  - local: protoc-gen-go
    out: gen
    opt: paths=source_relative
  - local: protoc-gen-go-grpc
    out: gen
    opt: paths=source_relative
EOF

echo "Generating Go code with buf..."
(
  cd "${WORK_DIR}/sui-apis"
  PATH="$(go env GOPATH)/bin:${PATH}" go run "github.com/bufbuild/buf/cmd/buf@${BUF_VERSION}" \
    generate --path proto/sui/rpc/v2
)

echo "Vendoring generated files into ${OUT_DIR}..."
rm -f "${OUT_DIR}"/*.pb.go
mkdir -p "${OUT_DIR}"
cp "${WORK_DIR}/sui-apis/gen/sui/rpc/v2/"*.go "${OUT_DIR}/"

echo "Verifying no deprecated golang/protobuf imports..."
if grep -rl "github.com/golang/protobuf" "${OUT_DIR}" >/dev/null 2>&1; then
  echo "ERROR: generated code imports deprecated github.com/golang/protobuf" >&2
  exit 1
fi

echo "Building..."
(cd "${REPO_ROOT}" && go build ./apps/chain_pusher/pkg/sui/...)

echo "Done. Generated $(ls "${OUT_DIR}"/*.pb.go | wc -l | tr -d ' ') files."
