# Stage 1: Build stage
FROM golang:1.22-alpine AS build

# Set the working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download -x

# Copy the source code
COPY . ./

# Build the Go application
RUN go build -v -o ./healthcheck ./cmd/main.go

# Stage 2: Final stage
FROM alpine:edge

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/healthcheck .

# Set Necessary Environment Variables needed for the application
ENV APP_ENV=test
ENV POSTGRES_HOST=host.docker.internal
ENV POSTGRES_PORT=5432
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=mysecretpassword
ENV POSTGRES_DB=healthcheck
ENV WEBHOOK_URL=http://localhost:8082/webhook

# Set the entrypoint command
ENTRYPOINT ["./healthcheck"]
