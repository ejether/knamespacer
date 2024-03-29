FROM golang:1.22
ARG TARGETOS
ARG TARGETARCH

RUN apt update && apt upgrade -y && apt install dumb-init -y
RUN go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest && setup-envtest use

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
