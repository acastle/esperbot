FROM golang:1.15.6 as builder

COPY go.mod .
RUN go get

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/acastle/esperbot/app .
CMD ["./app"]
