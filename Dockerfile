ARG GOLANG_VERSION=1.11

FROM golang:${GOLANG_VERSION} as builder
WORKDIR /go/src/github.com/b-b3rn4rd/4-state-barcode/

ARG IMAGE_TAG=1.0.0
COPY . ./
RUN go get -u github.com/golang/dep/cmd/dep \
    && dep ensure \
#    && go get gopkg.in/alecthomas/gometalinter.v2 \
#    && gometalinter.v2 --install \
#    && gometalinter.v2 ./... \
    && go test -v ./...

RUN CGO_ENABLED=0 \
        GOOS=linux \
        go build \
        -ldflags "-X main.Version=${IMAGE_TAG}" \
        -a -o app cli/main.go

FROM alpine:latest
WORKDIR /home/appuser
ENTRYPOINT ["./app"]
CMD ["--help"]

RUN addgroup -g 1000 appuser && \
        adduser -D -u 1000 -G appuser appuser -h /home/appuser

COPY --chown=appuser:appuser --from=builder /go/src/github.com/b-b3rn4rd/4-state-barcode/app .
USER appuser
