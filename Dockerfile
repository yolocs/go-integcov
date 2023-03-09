FROM --platform=$BUILDPLATFORM golang:1.20 AS builder

FROM cgr.dev/chainguard/static

COPY go-integcov /go-integcov
ENTRYPOINT ["/go-integcov"]
