ARG GOLANG_VERSION=1.11

FROM golang:${GOLANG_VERSION} as builder
WORKDIR /go/src/github.com/b-b3rn4rd/4-state-barcode/

ARG IMAGE_TAG=1.0.0
COPY . ./
RUN apt-get -y update \
    && apt-get install -y zip \
    && go get -u github.com/golang/dep/cmd/dep \
    && dep ensure \
    && go get gopkg.in/alecthomas/gometalinter.v2 \
    && gometalinter.v2 --install \
    && gometalinter.v2 ./... \
    && go test -v ./...

RUN CGO_ENABLED=0 \
        GOOS=linux \
        go build \
        -ldflags "-X main.Version=${IMAGE_TAG}" \
        -a -o app-linux cli/main.go  \
    && mkdir -p /tmp/release/linux \
    && cp app-linux /tmp/release/linux/4-state-barcode
RUN CGO_ENABLED=0 \
        GOOS=darwin \
        go build \
        -ldflags "-X main.Version=${IMAGE_TAG}" \
        -a -o app-darwin cli/main.go \
    && mkdir -p /tmp/release/darwin \
    && cp app-darwin /tmp/release/darwin/4-state-barcode
RUN CGO_ENABLED=0 \
        GOOS=windows \
        go build \
        -ldflags "-X main.Version=${IMAGE_TAG}" \
        -a -o app-windows cli/main.go \
    && mkdir -p /tmp/release/windows \
    && cp app-windows /tmp/release/windows/4-state-barcode
RUN cd /tmp/release && zip -r /tmp/release.zip *
FROM alpine:latest
WORKDIR /home/appuser
ENTRYPOINT ["./app-linux"]
CMD ["--help"]

RUN addgroup -g 1000 appuser && \
        adduser -D -u 1000 -G appuser appuser -h /home/appuser

COPY --chown=appuser:appuser --from=builder /go/src/github.com/b-b3rn4rd/4-state-barcode/app-linux .
COPY --chown=appuser:appuser --from=builder /tmp/release.zip /tmp/release.zip
USER appuser
