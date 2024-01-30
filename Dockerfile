FROM golang:1.21-alpine as builder

ARG GITHUB_REF
ARG GITHUB_SHA

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o bin/pidor-bot

FROM alpine:3.19

COPY --from=builder /app/bin/pidor-bot /bin/pidor-bot

CMD ["/bin/pidor-bot", "run"]