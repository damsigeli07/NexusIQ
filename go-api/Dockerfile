#  Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -mod=vendor -o server .

FROM golang:1.25-alpine
WORKDIR /app
COPY --from=builder /app/server .
RUN chmod +x /app/server
EXPOSE 8080
CMD ["./server"]