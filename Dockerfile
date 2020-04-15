# build image
FROM golang:1.14.2-alpine AS build-env

# install build tools
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# build
WORKDIR /app
# copy sources
COPY . .
# setup build env
WORKDIR /app/cli
# vendor build only can be executed outside the GOPATH
RUN go build -mod=vendor .

# distribution image
FROM alpine:3.11

# add CAs
RUN apk --no-cache add ca-certificates

COPY --from=build-env /app/cli/cli .

# start
CMD ["./cli", "start"]
