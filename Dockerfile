FROM golang:alpine AS build

# Install build tools
RUN apk --no-cache add gcc g++ make git

# Set build env variables
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .

RUN go build -ldflags "-s -w" -o refractor-bin ./main.go

# Create actual container
FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /var/refractor

RUN mkdir ./auth
RUN mkdir ./auth/templates
RUN mkdir ./auth/static
RUN mkdir ./internal
RUN mkdir ./internal/mail
RUN mkdir ./internal/mail/templates

# Copy the binary from the build stage into /var/refractor
COPY --from=build /build/refractor-bin ./refractor
COPY --from=build /build/auth/templates ./auth/templates
COPY --from=build /build/auth/static ./auth/static
COPY --from=build /build/internal/mail/templates ./internal/mail/templates

ENTRYPOINT PORT=80 /var/refractor/refractor