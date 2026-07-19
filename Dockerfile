ARG UBUNTU_VERSION=26.04
ARG GO_VERSION=1.26.5
ARG GO_LINUX_AMD64_SHA256=5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053

FROM ubuntu:${UBUNTU_VERSION} AS build-base
ARG GO_VERSION
ARG GO_LINUX_AMD64_SHA256
ARG TARGETARCH=amd64

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=/usr/local/go/bin:/root/go/bin:$PATH

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl git build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" -o /tmp/go.tgz \
    && echo "${GO_LINUX_AMD64_SHA256}  /tmp/go.tgz" | sha256sum -c - \
    && tar -C /usr/local -xzf /tmp/go.tgz \
    && rm -f /tmp/go.tgz \
    && go version

WORKDIR /src
COPY go.mod ./
COPY . .

FROM build-base AS test
RUN go test -race -cover ./...
RUN go vet ./...
RUN test -z "$(gofmt -l .)"

FROM test AS build
ARG VERSION=dev
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -trimpath \
    -tags netgo \
    -ldflags "-s -w -X main.VERSION=${VERSION}" \
    -o /out/hosts-file-updater .

FROM ubuntu:${UBUNTU_VERSION} AS runtime
ARG VERSION=dev

# This legacy-platform sidekick must run as root to rewrite Docker-managed
# /etc/hosts. See SECURITY.md before changing the runtime user.
LABEL org.opencontainers.image.title="hosts-file-updater" \
      org.opencontainers.image.description="Metadata-driven hosts file updater maintained by PastureStack" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/hosts-file-updater" \
      org.opencontainers.image.vendor="PastureStack community" \
      io.pasturestack.compatibility.contract="legacy-metadata-v2015-12-19"

COPY --from=build /out/hosts-file-updater /usr/bin/hosts-file-updater
RUN rm -f /usr/bin/pebble \
    && ln -s hosts-file-updater /usr/bin/etc-host-updater
COPY LICENSE NOTICE ORIGIN.md COMPATIBILITY.md /licenses/

CMD ["hosts-file-updater", "--debug"]
