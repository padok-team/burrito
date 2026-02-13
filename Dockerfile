# syntax=docker/dockerfile:1.21.0@sha256:27f9262d43452075f3c410287a2c43f5ef1bf7ec2bb06e8c9eeb1b8d453087bc

# Build Burrito UI

FROM docker.io/library/node:24.13.1@sha256:00e9195ebd49985a6da8921f419978d85dfe354589755192dc090425ce4da2f7 AS builder-ui

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
FROM docker.io/library/golang:1.26.0-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS builder
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

FROM docker.io/library/alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

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
RUN mkdir -p /var/run/burrito/repositories
RUN chmod +x /usr/local/bin/*
# /home/burrito/.config is required for debug mode
RUN mkdir -p /home/burrito/.config && chown -R burrito:burrito /runner /home/burrito /var/run/burrito

# Use an unprivileged user
USER 65532:65532

# Run Burrito on container startup
ENTRYPOINT ["burrito"]
