FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS base

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /home/app/svc-http-log-server
RUN --mount=type=bind,source=./svc-http-log-server,target=/home/app/svc-http-log-server \
    go mod download && GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o /home/app/bin/svc-http-log-server .

ENTRYPOINT [ "/home/app/bin/svc-http-log-server" ]
