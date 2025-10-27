FROM golang:latest AS builder
WORKDIR /work

ARG VERSION

COPY ./go.mod ./go.sum ./
RUN go mod download && go mod verify
COPY ./ ./
RUN VERSION=${VERSION:-$(git describe)} \
    BUILD_MACHINE=$(uname -srmo) \
    BUILD_TIME=$(date) \
    GO_VERSION=$(go version) \
    go build -ldflags "-s -w -X main.version=${VERSION} -X \"main.buildMachine=${BUILD_MACHINE}\" -X \"main.buildTime=${BUILD_TIME}\" -X \"main.goVersion=${GO_VERSION}\"" -o yisp

FROM gcr.io/distroless/base:nonroot
WORKDIR /work

ENV HOME=/home/nonroot

USER nonroot:nonroot
COPY --from=builder /work/yisp /yisp
COPY --chown=nonroot:nonroot ./cache /home/nonroot/.cache/yisp

ENTRYPOINT ["/yisp"]
CMD ["krm"]
