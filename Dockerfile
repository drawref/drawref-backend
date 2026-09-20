## build drawref binary
FROM --platform=$BUILDPLATFORM docker.io/golang:1.27-alpine AS build-env

WORKDIR /go/src/github.com/drawref/drawref-backend

# cache dependencies (invalidated only when go.mod/go.sum change)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 2. Copy source code
COPY . .

# compile natively for the target platform using Go cross-compilation
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT=""
ARG GIT_TAG=""

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -v \
    -ldflags "-X main.commit=${GIT_COMMIT} -X main.version=${GIT_TAG}" \
    -o /go/bin/drawref-backend .

## build drawref container
FROM docker.io/alpine:3.19

LABEL maintainer="Daniel Oaks <daniel@danieloaks.net>" \
      description="Drawref is a webapp that holds and presents images for drawing reference"

EXPOSE 8465/tcp

COPY --from=build-env /go/bin/drawref-backend \
                      /go/src/github.com/drawref/drawref-backend/distrib/docker/run.sh \
                      /drawref-bin/
COPY --from=build-env /go/src/github.com/drawref/drawref-backend/migrations \
                      /drawref-bin/migrations

ENTRYPOINT ["/drawref-bin/run.sh"]

# # uncomment to debug
# RUN apk add --no-cache bash
# RUN apk add --no-cache vim
# CMD /bin/bash
