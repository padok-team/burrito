# Build the manager binary
FROM docker.io/library/golang:1.19.9@sha256:8cefba2710250b21a8b8e32281788c5b53dc561ba0c51ea7de92b9a350663b7d as builder
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

# Use an unprivileged user
USER 65532:65532

# Run Burrito on container startup
ENTRYPOINT ["burrito"]
