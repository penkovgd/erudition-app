FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/knowledge

RUN go mod download

RUN CGO_ENABLED=0 go build -o /server ./cmd/main.go

FROM alpine:3.20

COPY --from=builder /server /server

ENTRYPOINT [ "/server" ]