# Build and run hsmdemo against SoftHSM in a container.
#
#   docker build -t hsmdemo .
#   docker run --rm -e HSM_PIN=1234 -e HSM_SO_PIN=5678 hsmdemo
#
# The token is created on first start under /var/lib/hsmdemo/tokens. Mount a
# named volume there to keep the key across runs.
#
# Both base images are pinned by digest: bump them deliberately, never by
# silently retagging.

# --- build ----------------------------------------------------------------
FROM golang:1.27.2-alpine3.24@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS build

# miekg/pkcs11 is a cgo wrapper around the PKCS#11 C API, so CGO stays on.
# Alpine needs gcc and musl-dev for that; the -ldl it links is a musl stub.
RUN apk add --no-cache gcc musl-dev

WORKDIR /src
ENV CGO_ENABLED=1

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -trimpath -ldflags='-s -w' -o /out/hsmdemo .

# --- runtime --------------------------------------------------------------
FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

# Provides /usr/lib/softhsm/libsofthsm2.so and softhsm2-util.
RUN apk add --no-cache softhsm

ENV PKCS11_LIB_PATH=/usr/lib/softhsm/libsofthsm2.so \
    SOFTHSM2_CONF=/etc/hsmdemo/softhsm2.conf

# Unprivileged runtime user. The token store must be writable by it, because
# the token is created on first start, not baked into the image.
RUN addgroup -S -g 10001 app \
 && adduser -S -u 10001 -G app -h /var/lib/hsmdemo app \
 && install -d -m 0755 /etc/hsmdemo \
 && printf '%s\n' \
      '# Generated during the image build; tokendir is the volume mount point.' \
      'directories.tokendir = /var/lib/hsmdemo/tokens' \
      'objectstore.backend = file' \
      'objectstore.umask = 0077' \
      'log.level = INFO' \
      > /etc/hsmdemo/softhsm2.conf \
 && install -d -m 0700 -o app -g app /var/lib/hsmdemo/tokens

COPY --from=build /out/hsmdemo /usr/local/bin/hsmdemo
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod 0755 /usr/local/bin/docker-entrypoint.sh

USER app
VOLUME ["/var/lib/hsmdemo/tokens"]

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
