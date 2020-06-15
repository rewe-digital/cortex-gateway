
# build image
FROM golang:1.14-alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -installsuffix cgo -o /go/bin/cortex-gateway

# executable image
FROM alpine:3
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/cortex-gateway /go/bin/cortex-gateway

ENV VERSION 0.1.0
ENTRYPOINT ["/go/bin/cortex-gateway"]
