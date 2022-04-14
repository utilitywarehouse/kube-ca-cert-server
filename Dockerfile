FROM golang:1.18-alpine AS build
WORKDIR /go/src/github.com/utilitywarehouse/kube-ca-cert-server
COPY . /go/src/github.com/utilitywarehouse/kube-ca-cert-server
ENV CGO_ENABLED 0
RUN \
  apk --no-cache add git upx \
    && go get -t ./... \
    && go test -v \
    && go build -ldflags='-s -w' -o /kube-ca-cert-server . \
    && upx /kube-ca-cert-server

FROM alpine:3.15
COPY --from=build /kube-ca-cert-server /kube-ca-cert-server

ENTRYPOINT [ "/kube-ca-cert-server" ]
