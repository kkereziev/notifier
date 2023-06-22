FROM golang:1.20-alpine3.17 as base

ENV BASE_DIR /go/src/github.com/kkereziev/notifier

WORKDIR ${BASE_DIR}

COPY go.mod go.sum ${BASE_DIR}/

RUN go mod download -x

COPY cmd ${BASE_DIR}/cmd
COPY internal ${BASE_DIR}/internal

FROM base as dev

RUN go install github.com/githubnemo/CompileDaemon@v1.4.0

FROM base as builder

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /dist/server ./cmd/server/main.go \
    && CGO_ENABLED=0 GOOS=linux go build -v -o /dist/mail-cron ./cmd/mail-cron/main.go \
    && CGO_ENABLED=0 GOOS=linux go build -v -o /dist/slack-cron ./cmd/slack-cron/main.go \
    && CGO_ENABLED=0 GOOS=linux go build -v -o /dist/sms-cron ./cmd/sms-cron/main.go

FROM alpine:3.16 as prod

RUN apk update && apk add tzdata

ENV BASE_DIR /go/src/github.com/kkereziev/notifier

COPY --from=builder /dist .

EXPOSE 8000

CMD ["/server"]
