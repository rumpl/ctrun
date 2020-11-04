# syntax=docker/dockerfile:experimental

ARG GO_VERSION=1.15.0-alpine
ARG GOLANGCI_LINT_VERSION=v1.30.0-alpine

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS base
WORKDIR /ctrun
ENV GO111MODULE=on
RUN apk add --no-cache make
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM base AS make-cli
ENV CGO_ENABLED=0
ARG TARGETOS
ARG TARGETARCH
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    make BINARY=/out/docker -f builder.Makefile cli

FROM scratch AS cli
COPY --from=make-cli /out/* .
