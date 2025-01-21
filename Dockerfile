FROM node:20-alpine as frontend-builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY src/ ./src/
COPY tailwind.config.js ./
RUN npm run build:css

FROM golang:1.21-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/static/css/styles.css ./static/css/
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main /app/
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/.env.example /app/.env

EXPOSE 8080
CMD ["/app/main"]
