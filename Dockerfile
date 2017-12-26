FROM golang:1.9.2 as builder
WORKDIR /go/src/github.com/acastle/esperbot
COPY . .
RUN go get \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/acastle/esperbot/app .
CMD ["./app"]
