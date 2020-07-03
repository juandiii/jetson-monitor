FROM golang:1.14-alpine as BUILDER

WORKDIR /go/src/jetson-monitor
COPY . .
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/jetson-monitor

FROM alpine:3.12.0

RUN apk update && \
    apk add ca-certificates \
    rm -rf /var/cache/apk/* && \
    adduser -D -g '' -h /var/jetson-monitor jetson-monitor

VOLUME /var/jetson-monitor
WORKDIR /var/jetson-monitor

COPY --from=BUILDER /go/bin/jetson-monitor /bin/
RUN chmod +x /bin/jetson-monitor

USER jetson-monitor
EXPOSE 38080

CMD ["jetson-monitor"]
