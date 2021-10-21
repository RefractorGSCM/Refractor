FROM golang:alpine AS build

# Install build tools
RUN apk --no-cache add gcc g++ make git

# Set build env variables
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY src/go.mod .
COPY src/go.sum .
RUN go mod download

COPY /.git .
COPY /src .

RUN go build -ldflags "-s -w -X main.VERSION=`echo $(git rev-parse --abbrev-ref HEAD):$(git rev-parse --short HEAD)`" -o refractor-bin main.go

# Create actual container
FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /var/refractor

# Create directories where folders containing static assets are located. This is likely not the ideal way to do this.
# However, it works fine! If many more static assets are added, a more scalable solution would be warranted.
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
