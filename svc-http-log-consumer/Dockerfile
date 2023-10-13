FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS base

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /home/app/svc-http-log-consumer
RUN --mount=type=bind,source=./svc-http-log-consumer,target=/home/app/svc-http-log-consumer \
    go mod download && GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o /home/app/bin/svc-http-log-consumer .

ENTRYPOINT [ "/home/app/bin/svc-http-log-consumer" ]
