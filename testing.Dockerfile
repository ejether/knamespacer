FROM golang:1.21
ARG TARGETOS
ARG TARGETARCH

RUN go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
RUN setup-envtest use

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download