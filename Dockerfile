# syntax=docker/dockerfile:1.3
FROM golang:1.16-alpine AS build

ARG GO_VERSION=1.16
ARG GORELEASER_XX_VERSION=1.2.5

FROM --platform=$BUILDPLATFORM crazymax/goreleaser-xx:${GORELEASER_XX_VERSION} AS goreleaser-xx
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS base
COPY --from=goreleaser-xx / /
RUN apk add --no-cache file git
WORKDIR /go/src/github.com/docker/distribution

FROM base AS build
ENV GO111MODULE=auto
ENV CGO_ENABLED=0
ARG TARGETPLATFORM
ARG PKG="github.com/distribution/distribution"
ARG BUILDTAGS="include_oss include_gcs"
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=6
ARG VERSION
ARG REVISION
RUN --mount=type=bind,rw \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=target=/go/pkg/mod,type=cache \
  goreleaser-xx --debug \
    --name="registry" \
    --dist="/out" \
    --main="./cmd/registry" \
    --flags="-v" \
    --ldflags="-s -w -X '$PKG/version.Version={{.Version}}' -X '$PKG/version.Revision={{.Commit}}' -X '$PKG/version.Package=$PKG'" \
    --tags="$BUILDTAGS" \
    --files="LICENSE" \
    --files="README.md"

FROM scratch AS artifacts
COPY --from=build /out/*.tar.gz /
COPY --from=build /out/*.zip /
COPY --from=build /out/*.sha256 /

FROM scratch AS binary
COPY --from=build /usr/local/bin/registry* /

FROM alpine:3.14
RUN apk add --no-cache ca-certificates
COPY cmd/registry/config-dev.yml /etc/docker/registry/config.yml
COPY --from=build /usr/local/bin/registry /bin/registry
VOLUME ["/var/lib/registry"]
EXPOSE 5000
ENTRYPOINT ["registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]
