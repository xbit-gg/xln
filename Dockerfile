FROM golang:1.16-alpine as builder

ARG checkout="master"

COPY . /build

RUN apk add --no-cache --update alpine-sdk \
        git \
        make \
        gcc \
        protobuf-dev \
        && go get github.com/golang/protobuf/protoc-gen-go \
        && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN cd /build \
&&  git checkout "$checkout" \
&&  make build

FROM alpine as final

# Data persistence
VOLUME /root/.xln

# Add utilities for quality of life and SSL-related reasons.
RUN apk --no-cache add \
    bash \
    curl

# Copy the binary from the builder image.
COPY --from=builder /build/xln /bin/

EXPOSE 5550 5551

ENTRYPOINT ["/bin/xln"]