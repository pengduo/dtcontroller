# Build the manager binary
FROM golang:1.16 as builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
# COPY dubbo/ dubbo/
COPY api/ api/
COPY controllers/ controllers/
COPY util/ util/
COPY vmsdk/ vmsdk/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

FROM alpine:latest
# Set env 
# ENV APP_LOG_CONF_FILE /log.yml
# ENV CONF_PROVIDER_FILE_PATH  /server.yml
# ENV CONF_CONSUMER_FILE_PATH /client.yml
WORKDIR /
COPY --from=builder /workspace/manager .
# COPY dubbo/log.yml .
# COPY dubbo/server.yml .
# COPY dubbo/client.yml .
USER 65532:65532

ENTRYPOINT ["/manager"]
