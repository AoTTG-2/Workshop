# Build
FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -o /workshop \
    ./cmd/workshop/main.go

# Deploy
FROM alpine

RUN apk --no-cache add ca-certificates

COPY --from=builder /workshop /bin/workshop
COPY ./migrations/postgres /migrations/postgres

ENTRYPOINT ["/bin/workshop"]
