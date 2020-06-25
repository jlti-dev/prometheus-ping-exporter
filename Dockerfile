FROM golang:latest as builder
WORKDIR /app
RUN go get github.com/tatsushid/go-fastping && \
    go get github.com/prometheus/client_golang/prometheus && \
    go get github.com/prometheus/client_golang/prometheus/promauto && \
    go get github.com/prometheus/client_golang/prometheus/promhttp
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
WORKDIR /app
Copy start.sh /app/start.sh
COPY --from=builder /app .
CMD ["/bin/sh", "start.sh"]
