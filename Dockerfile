FROM golang:1.23.1-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download && \
    go mod verify && \
    CGO_ENABLED=0 go build -o ./bin/app ./src/cmd/app/main.go

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/bin/app /app