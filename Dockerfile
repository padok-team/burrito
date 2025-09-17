# syntax=docker/dockerfile:1.18.0

# Build Burrito UI

FROM docker.io/library/node:22.19.0@sha256:afff6d8c97964a438d2e6a9c96509367e45d8bf93f790ad561a1eaea926303d9 AS builder-ui

WORKDIR /workspace
# Copy the node modules manifests
COPY ui/package.json ui/yarn.lock ./
# Install build dependencies
RUN yarn install --frozen-lockfile

# Copy the UI source
COPY ui .
# Set the API base URL
ENV VITE_API_BASE_URL=/api
RUN yarn build

# Build the manager binary
FROM docker.io/library/golang:1.24.7-alpine@sha256:fc2cff6625f3c1c92e6c85938ac5bd09034ad0d4bc2dfb08278020b68540dbb5 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG PACKAGE=github.com/padok-team/burrito
ARG COMMIT_HASH
ARG BUILD_TIMESTAMP
ARG BUILD_MODE=Release

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY internal/ internal/
COPY cmd/ cmd/
# Copy the UI build artifacts
COPY --from=builder-ui /workspace/dist internal/server/dist

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
ARG VERSION
ENV GOCACHE=/root/.cache/go-build

RUN if [ "${BUILD_MODE}" = "Debug" ]; then go install github.com/go-delve/delve/cmd/dlv@latest; fi

# Build with different flags based on debug mode
RUN --mount=type=cache,target=/root/.cache/go-build \
    if [ "${BUILD_MODE}" = "Debug" ]; then \
        GCFLAGS="all=-N -l"; \
        LDFLAGS=""; \
    else \
        GCFLAGS=""; \
        LDFLAGS="-w -s"; \
    fi && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build \
        -gcflags "${GCFLAGS}" \
        -ldflags="${LDFLAGS} -X ${PACKAGE}/internal/version.Version=${VERSION} \
        -X ${PACKAGE}/internal/version.CommitHash=${COMMIT_HASH} \
        -X ${PACKAGE}/internal/version.BuildTimestamp=${BUILD_TIMESTAMP}" \
        -o bin/burrito main.go

FROM docker.io/library/alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

WORKDIR /home/burrito

# Install required packages
RUN apk add --update --no-cache git bash openssh

ENV UID=65532
ENV GID=65532
ENV USER=burrito
ENV GROUP=burrito

# Create a non-root user to run the app
RUN addgroup \
  -g $GID \
  $GROUP && \
  adduser \
  --disabled-password \
  --no-create-home \
  --home $(pwd) \
  --uid $UID \
  --ingroup $GROUP \
  $USER

# Copy the binary to the production image from the builder stage
# Copy /go/bin/dlv*: the wildcard makes the copy to work, even if the binary is not present (in Release mode)
COPY --from=builder /workspace/bin/burrito /go/bin/dlv* /usr/local/bin/

RUN mkdir -p /runner/bin
RUN chmod +x /usr/local/bin/*
# /home/burrito/.config is required for debug mode
RUN mkdir -p /home/burrito/.config && chown -R burrito:burrito /runner /home/burrito

# Use an unprivileged user
USER 65532:65532

# Run Burrito on container startup
ENTRYPOINT ["burrito"]
