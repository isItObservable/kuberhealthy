FROM golang:1.19.2-alpine AS builder
RUN apk add build-base protobuf-dev protoc
WORKDIR /usr/src/app/

# Restore dependencies
COPY ./traceTest/ ./
RUN apk --no-cache add git
RUN go get github.com/isItObservable/kuberhealthy/tree/master/traceTest
RUN go mod download
# Build executable
RUN go build -o /go/bin/traceTest/ ./

# -----------------------------------------------------------------------------

FROM alpine

WORKDIR /usr/src/app/
RUN apk --no-cache add curl bash
RUN curl -L https://raw.githubusercontent.com/kubeshop/tracetest/main/install-cli.sh | bash
COPY --from=builder /go/bin/traceTest/ ./

ENTRYPOINT [ "./traceTest" ]


