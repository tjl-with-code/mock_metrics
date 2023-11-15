# Build Stage
FROM golang:1.17 AS build-stage

WORKDIR /app

RUN go env -w GO111MODULE=on && go env -w GOPROXY="https://goproxy.cn,direct"

COPY . .

RUN make build

# Runtime Stage
FROM alpine:3
COPY --from=build-stage /app/bin/fake-matrics /app/fake-matrics
COPY ./static /app/static

ENV GIN_MODE=release


WORKDIR /app
EXPOSE 8080

ENTRYPOINT ["/app/fake-matrics"]