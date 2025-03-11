## build drawref binary
FROM docker.io/golang:1.23-alpine AS build-env

RUN apk upgrade -U --force-refresh --no-cache && apk add --no-cache --purge --clean-protected -l -u make git

# copy drawref source
WORKDIR /go/src/github.com/drawref/drawref-backend
COPY . .

# compile
RUN make install

## build drawref container
FROM docker.io/alpine:3.19

# metadata
LABEL maintainer="Daniel Oaks <daniel@danieloaks.net>" \
      description="Drawref is a webapp that holds and presents images for drawing reference"

# standard port listened on
EXPOSE 8465/tcp

# drawref itself
COPY --from=build-env /go/bin/drawref-backend \
                      /go/src/github.com/drawref/drawref-backend/distrib/docker/run.sh \
                      /drawref-bin/
COPY --from=build-env /go/src/github.com/drawref/drawref-backend/migrations \
                      /drawref-bin/migrations

# launch
ENTRYPOINT ["/drawref-bin/run.sh"]

# # uncomment to debug
# RUN apk add --no-cache bash
# RUN apk add --no-cache vim
# CMD /bin/bash
