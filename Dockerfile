FROM golang:1.17 as builder
WORKDIR /app
COPY app/ .
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
WORKDIR /app
COPY start.sh /app/start.sh
COPY --from=builder /app .
EXPOSE 8080
CMD ["/bin/sh", "start.sh"]
