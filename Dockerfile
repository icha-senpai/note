FROM --platform=$BUILDPLATFORM node:22 AS node-build

ARG NPM_REGISTRY=

WORKDIR /app
ADD app/package.json app/pnpm* app/.npmrc .

RUN <<EORUN
#!/bin/bash -e
corepack enable
corepack install --global $(node -e 'console.log(require("./package.json").packageManager)')
npm config set registry ${NPM_REGISTRY}
pnpm install --silent
EORUN

ADD app/ .
RUN <<EORUN
#!/bin/bash -e
pnpm run build
node scripts/trimChangelogs.js
mkdir /artifacts
mv appearance stage guide /artifacts/
if [ -d changelogs ]; then mv changelogs /artifacts/; fi
EORUN

FROM golang:1.26-alpine AS go-build

RUN <<EORUN
#!/bin/sh -e
apk add --no-cache gcc musl-dev
go env -w GO111MODULE=on
go env -w CGO_ENABLED=1
EORUN

WORKDIR /kernel
ADD kernel/go.* .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg \
    go mod download

ADD kernel/ .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg \
    go build -tags fts5 -ldflags "-s -w"

FROM alpine:latest
LABEL maintainer="Scribli contributors"

RUN apk add --no-cache ca-certificates tzdata su-exec

ENV TZ=UTC
ENV HOME=/home/scribli
ENV RUN_IN_CONTAINER=true
EXPOSE 6806

WORKDIR /opt/scribli/
COPY --from=go-build --chmod=755 /kernel/kernel /kernel/entrypoint.sh .
COPY --from=node-build /artifacts .

ENTRYPOINT ["/opt/scribli/entrypoint.sh"]
# Start the server by default. Extra `docker run` / `command:` args replace CMD,
# so callers must include the `serve` subcommand when overriding it.
CMD ["serve"]
