FROM golang:1.22 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy in the repo
COPY . /knamespacer
WORKDIR /knamespacer

RUN make build


FROM alpine:3.19

LABEL org.opencontainers.image.licenses="Apache License 2.0"
WORKDIR /
# 'nobody' user in alpine
USER 65534:65534
COPY --from=builder /knamespacer/knamespacer .
ENTRYPOINT ["/knamespacer"]
