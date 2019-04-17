FROM golang:1.12

WORKDIR /app
RUN mkdir /build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/server ./cmd/manager

FROM alpine:latest

RUN apk update && \
   apk add ca-certificates && \
   update-ca-certificates && \
   rm -rf /var/cache/apk/*

COPY --from=0 /build /app
WORKDIR /app

EXPOSE 80

ENTRYPOINT ["/app/server"]