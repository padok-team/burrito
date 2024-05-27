# Build Burrito UI

FROM docker.io/library/node:20.11.1@sha256:e06aae17c40c7a6b5296ca6f942a02e6737ae61bbbf3e2158624bb0f887991b5 AS builder-ui

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
FROM docker.io/library/golang:1.22.3@sha256:f43c6f049f04cbbaeb28f0aad3eea15274a7d0a7899a617d0037aec48d7ab010 as builder
ARG TARGETOS
ARG TARGETARCH
ARG PACKAGE=github.com/padok-team/burrito
ARG VERSION
ARG COMMIT_HASH
ARG BUILD_TIMESTAMP

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
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a \
  -ldflags="\
  -X ${PACKAGE}/internal/version.Version=${VERSION} \
  -X ${PACKAGE}/internal/version.CommitHash=${COMMIT_HASH} \
  -X ${PACKAGE}/internal/version.BuildTimestamp=${BUILD_TIMESTAMP}" \
  -o bin/burrito main.go

FROM docker.io/library/alpine:3.20.0@sha256:77726ef6b57ddf65bb551896826ec38bc3e53f75cdde31354fbffb4f25238ebd

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
COPY --from=builder /workspace/bin/burrito /usr/local/bin/burrito

RUN mkdir -p /runner/bin
RUN chmod +x /usr/local/bin/burrito
RUN chown -R burrito:burrito /runner

# Use an unprivileged user
USER 65532:65532

# Run Burrito on container startup
ENTRYPOINT ["burrito"]
