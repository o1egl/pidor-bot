FROM golang:1.18-alpine as builder

RUN apk add --no-cache git curl

WORKDIR /app

COPY scripts/ scripts/
RUN scripts/task_install.sh -d -b /usr/local/bin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN task build

FROM alpine:3.16

COPY --from=builder /app/bin/pidor-bot /bin/pidor-bot

CMD ["/bin/pidor-bot", "run"]