FROM golang:1.20-alpine3.17 as base

ENV BASE_DIR /go/src/github.com/kkereziev/notifier

WORKDIR ${BASE_DIR}

COPY go.mod go.sum ${BASE_DIR}/

RUN go mod download -x

COPY internal ${BASE_DIR}/internal
COPY main.go ${BASE_DIR}/main.go

FROM base as dev

RUN go install github.com/githubnemo/CompileDaemon@v1.4.0

FROM base as builder

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /dist/server ./main.go

FROM alpine:3.16 as prod

RUN apk update && apk add tzdata

ENV BASE_DIR /go/src/github.com/kkereziev/notifier

COPY --from=builder /dist .

EXPOSE 8000

CMD ["/server"]
