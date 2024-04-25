# Start from a Node.js base image to run Yarn commands
FROM --platform=linux/arm64/v8 node:latest as ui-builder

WORKDIR /ui

# Copy the full ui folder
COPY ui .

# Install dependencies and build the UI
RUN yarn install && yarn build


# Start from a Golang base image
FROM --platform=linux/arm64/v8 golang:latest as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=arm64

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN rm -rf ui
COPY --from=ui-builder /ui/dist ./ui/dist
COPY --from=ui-builder /ui/embed.go ./ui/embed.go
RUN go mod download

# Build the Go app
RUN go build -o ./autopus ./cmd/cli

# Build all apps
RUN make build-apps

# Start a new stage from scratch
FROM --platform=linux/arm64/v8 alpine:latest

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/autopus .
COPY --from=builder /app/dist/apps ./dist/apps

# Command to run the executable
CMD ["./autopus", "start"]
