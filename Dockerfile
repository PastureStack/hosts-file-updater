# syntax=docker/dockerfile:1.7
ARG UBUNTU_IMAGE=ubuntu:26.04@sha256:2260313b31c8c011cd2eebe728008efac1b3982be73eb71348ea2648d2c0e09b
ARG GO_VERSION=1.27.3
ARG GO_LINUX_AMD64_SHA256=65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40

FROM ${UBUNTU_IMAGE} AS build-base
ARG GO_VERSION
ARG GO_LINUX_AMD64_SHA256
ARG TARGETARCH=amd64
ARG UBUNTU_SNAPSHOT=20260825T000000Z

ADD --checksum=sha256:5b6326f1932cfd7be57130561dc78192f09ed9f58bfbe6606ed0c50510c3b6cd \
    https://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20260816ubuntu0.26.04.1_all.deb \
    /tmp/ca-certificates.deb

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=/usr/local/go/bin:/root/go/bin:$PATH

COPY ubuntu-security-packages.lock /licenses/ubuntu-security-packages.lock
COPY scripts/install-ubuntu-security-updates /usr/local/libexec/pasturestack/install-ubuntu-security-updates
ADD --checksum=sha256:5b6326f1932cfd7be57130561dc78192f09ed9f58bfbe6606ed0c50510c3b6cd https://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20260816ubuntu0.26.04.1_all.deb /opt/pasturestack-ca-certificates.deb
RUN set -eux; \
    mkdir -p /tmp/ca-bootstrap /etc/ssl/certs; \
    dpkg-deb --extract /tmp/ca-certificates.deb /tmp/ca-bootstrap; \
    find /tmp/ca-bootstrap/usr/share/ca-certificates -type f -name '*.crt' | LC_ALL=C sort | while IFS= read -r certificate; do sed -e '$a\' "${certificate}"; done > /etc/ssl/certs/ca-certificates.crt; \
    rm -rf /tmp/ca-bootstrap /tmp/ca-certificates.deb; \
    rm -f /etc/apt/sources.list /etc/apt/sources.list.d/*.list /etc/apt/sources.list.d/*.sources; \
    printf 'Types: deb\nURIs: https://snapshot.ubuntu.com/ubuntu/%s\nSuites: resolute resolute-updates resolute-backports resolute-security\nComponents: main universe restricted multiverse\nSigned-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg\nSnapshot: no\n' "${UBUNTU_SNAPSHOT}" > /etc/apt/sources.list.d/pasturestack-snapshot.sources; \
    printf 'Acquire::Retries "5";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\nAcquire::https::Verify-Peer "true";\nAcquire::https::Verify-Host "true";\nAcquire::AllowInsecureRepositories "false";\nAPT::Get::AllowUnauthenticated "false";\n' > /etc/apt/apt.conf.d/80pasturestack-snapshot; \
    apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl git build-essential \
    && dpkg -i /opt/pasturestack-ca-certificates.deb \
    && update-ca-certificates --fresh \
    && test "$(dpkg-query -W -f='${Version}' ca-certificates)" = '20260816ubuntu0.26.04.1' \
    && test ! -e /usr/share/ca-certificates/mozilla/emSign_ECC_Root_CA_-_G3.crt \
    && rm -f /opt/pasturestack-ca-certificates.deb \
    && sh /usr/local/libexec/pasturestack/install-ubuntu-security-updates /licenses/ubuntu-security-packages.lock /etc/ssl/certs/ca-certificates.crt \
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

FROM ${UBUNTU_IMAGE} AS runtime
ARG VERSION=dev
ARG REVISION=unknown
COPY --from=build-base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ubuntu-security-packages.lock /licenses/ubuntu-security-packages.lock
COPY scripts/install-ubuntu-security-updates /usr/local/libexec/pasturestack/install-ubuntu-security-updates
RUN sh /usr/local/libexec/pasturestack/install-ubuntu-security-updates /licenses/ubuntu-security-packages.lock /etc/ssl/certs/ca-certificates.crt

# This legacy-platform sidekick must run as root to rewrite Docker-managed
# /etc/hosts. See SECURITY.md before changing the runtime user.
LABEL org.opencontainers.image.title="hosts-file-updater" \
      org.opencontainers.image.description="Metadata-driven hosts file updater maintained by PastureStack" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/hosts-file-updater" \
      org.opencontainers.image.vendor="PastureStack community" \
      io.pasturestack.compatibility.contract="legacy-metadata-v2015-12-19"

COPY --from=build /out/hosts-file-updater /usr/bin/hosts-file-updater
RUN rm -f /usr/bin/pebble \
    && ln -s hosts-file-updater /usr/bin/etc-host-updater
COPY LICENSE NOTICE ORIGIN.md COMPATIBILITY.md /licenses/

CMD ["hosts-file-updater", "--debug"]
