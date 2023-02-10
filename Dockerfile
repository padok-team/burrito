# Build the manager binary
FROM golang:1.19 as builder
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

FROM golang:alpine

RUN apk add --update git bash openssh

WORKDIR /
COPY --from=builder /workspace/bin/burrito .
USER 65532:65532

ENTRYPOINT ["/burrito"]
