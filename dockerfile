FROM golang:1.26.2-alpine AS builder

WORKDIR /app

RUN apk add --no-cache make

COPY . /app/

RUN make build

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/kwik-mq /app/

EXPOSE 10526

CMD ["./kwik-mq"]