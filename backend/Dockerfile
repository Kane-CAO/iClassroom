FROM golang:1.25-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/iclassroom-server ./cmd/server

FROM golang:1.25-alpine

RUN apk add --no-cache ca-certificates tzdata wget \
  && addgroup -S app \
  && adduser -S app -G app

WORKDIR /app

COPY --from=build /out/iclassroom-server /app/iclassroom-server

RUN mkdir -p /app/uploads \
  && chown -R app:app /app

USER app

EXPOSE 8080

CMD ["/app/iclassroom-server"]
