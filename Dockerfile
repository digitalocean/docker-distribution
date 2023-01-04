FROM golang:1.16-alpine AS build
ARG GO_VERSION=1.16.15


FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS base
COPY --from=goreleaser-xx / /
RUN apk add --no-cache file git
WORKDIR /go/src/github.com/docker/distribution
ENV DISTRIBUTION_DIR /go/src/github.com/docker/distribution
ENV BUILDTAGS include_oss include_gcs

FROM base AS build
ENV GO111MODULE=auto
ENV CGO_ENABLED=0
# GIT_REF is used by goreleaser-xx to handle the proper git ref when available.
# It will fallback to the working tree info if empty and use "git tag --points-at"
# or "git describe" to define the version info.
ARG GIT_REF
ARG TARGETPLATFORM
ARG PKG="github.com/distribution/distribution"
ARG BUILDTAGS="include_oss include_gcs"
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=6
ARG VERSION
ARG REVISION
RUN set -ex \
    && apk add --no-cache make git file

FROM scratch AS artifact
COPY --from=build /out/*.tar.gz /
COPY --from=build /out/*.zip /
COPY --from=build /out/*.sha256 /

FROM scratch AS binary
COPY --from=build /usr/local/bin/registry* /
WORKDIR $DISTRIBUTION_DIR
COPY . $DISTRIBUTION_DIR
RUN CGO_ENABLED=0 make PREFIX=/go clean binaries && file ./bin/registry | grep "statically linked"

FROM alpine:3.14
RUN set -ex \
    && apk add --no-cache ca-certificates apache2-utils
COPY cmd/registry/config-dev.yml /etc/docker/registry/config.yml
COPY --from=build /go/src/github.com/docker/distribution/bin/registry /bin/registry
VOLUME ["/var/lib/registry"]
EXPOSE 5000
ENTRYPOINT ["registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]
