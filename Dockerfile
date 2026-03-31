# ---------- Build stage ----------
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
# COPY go.sum ./ (uncomment when external dependencies are added)
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fleet-monitor ./main

# ---------- Runtime stage ----------
FROM alpine:3.19

RUN adduser -D -u 1000 appuser
WORKDIR /app

COPY --from=builder /fleet-monitor .
COPY devices.csv .

USER appuser
EXPOSE 6733

ENTRYPOINT ["./fleet-monitor"]
CMD ["--addr", "0.0.0.0:6733"]
