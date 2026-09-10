# Single-node wasmd chain with the Stork CosmWasm contract deployed, used by the
# CosmWasm chain pusher integration tests.

# The wasm artifact is platform independent, so compile it natively on the build host
# with the optimizer image for that architecture.
FROM --platform=$BUILDPLATFORM cosmwasm/optimizer:0.17.0 AS optimizer-amd64
FROM --platform=$BUILDPLATFORM cosmwasm/optimizer-arm64:0.17.0 AS optimizer-arm64

FROM optimizer-${BUILDARCH} AS contract

COPY chains/cosmwasm/contracts/stork /code
WORKDIR /code
RUN optimize.sh .

# Keep in sync with the github.com/CosmWasm/wasmd version in go.mod.
FROM cosmwasm/wasmd:v0.61.9

RUN apk add --no-cache jq

ARG STORK_PUBLIC_KEY
ENV STORK_PUBLIC_KEY=${STORK_PUBLIC_KEY}

COPY --from=contract /code/artifacts/stork_cw.wasm /opt/stork_cw.wasm
COPY docker/scripts/cosmwasm-docker-entrypoint.sh /opt/cosmwasm-docker-entrypoint.sh
RUN chmod +x /opt/cosmwasm-docker-entrypoint.sh

EXPOSE 26657

ENTRYPOINT [ "/opt/cosmwasm-docker-entrypoint.sh" ]
