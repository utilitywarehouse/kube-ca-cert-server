FROM golang:1.14-alpine AS build
WORKDIR /go/src/github.com/ffilippopoulos/kube-ca-cert-server
COPY . /go/src/github.com/ffilippopoulos/kube-ca-cert-server
ENV CGO_ENABLED 0
RUN go build -o /kube-ca-cert-server .

FROM alpine:3.11
COPY --from=build /kube-ca-cert-server /kube-ca-cert-server

ENTRYPOINT [ "/kube-ca-cert-server" ]
