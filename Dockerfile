# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/gosigfmt \
    ./cmd/gosigfmt

FROM scratch
COPY --from=builder /out/gosigfmt /gosigfmt
WORKDIR /work
USER 65532:65532
ENTRYPOINT ["/gosigfmt"]
