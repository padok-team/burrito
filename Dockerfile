# Build Burrito UI

FROM docker.io/library/node:20.9.0@sha256:5f21943fe97b24ae1740da6d7b9c56ac43fe3495acb47c1b232b0a352b02a25c AS builder-ui

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
FROM docker.io/library/golang:1.20.7@sha256:bc5f0b5e43282627279fe5262ae275fecb3d2eae3b33977a7fd200c7a760d6f1 as builder
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

FROM docker.io/library/alpine:3.18.2@sha256:82d1e9d7ed48a7523bdebc18cf6290bdb97b82302a8a9c27d4fe885949ea94d1

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

RUN chmod +x /usr/local/bin/burrito
RUN mkdir /repository && chown -R burrito:burrito /repository

# Use an unprivileged user
USER 65532:65532

# Run Burrito on container startup
ENTRYPOINT ["burrito"]
