# -------------------------------------------------------
# Build the go source
# -------------------------------------------------------
FROM golang:1.12.1 as go-builder

COPY config.yaml /go/bin/config.yaml
COPY /webapp /app
COPY /pkg $GOPATH/src/github.com/att-cloudnative-labs/template-api/pkg
COPY /genesis_config $GOPATH/src/github.com/att-cloudnative-labs/template-api/genesis_config

WORKDIR /app

RUN go get -d -v ./... &&\
    GIT_TERMINAL_PROMPT=1 GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -v --installsuffix cgo --ldflags="-s" -o /go/bin/app && \
    echo "Genesis API Binary Built"
    
# -------------------------------------------------------
# Add the Go binary to Alpine to create an enhanced base image
# -------------------------------------------------------
FROM alpine:3.8

# copy source files from go-builder stage into the scratch container
COPY --from=go-builder /go/bin/app /usr/bin/app
COPY --from=go-builder /go/bin/config.yaml /usr/bin/config.yaml

RUN apk update && apk add --no-cache git ca-certificates

# change working directory, so CMD is brief
WORKDIR /usr/bin/

# run the Go application
CMD ["app"]