# Source for qemu-user-static >7.0
FROM public.ecr.aws/ubuntu/ubuntu:lunar as qemu_binaries

# hadolint ignore=DL3008
RUN apt-get update -qq \
 && apt-get install -qqy --no-install-recommends qemu-user-static

RUN mkdir -p /binaries && \
    if [ "$(uname -m)" = "aarch64" ]; then \
        cp /usr/bin/qemu-arm-static /binaries/; \
        cp /usr/bin/qemu-x86_64-static /binaries/; \
    else \
        cp /usr/bin/qemu-aarch64-static /binaries/; \
        cp /usr/bin/qemu-arm-static /binaries/; \
    fi

FROM golang:1.22-bullseye AS builder

# hadolint ignore=DL3008
RUN apt-get update -qq \
 && apt-get install -qqy --no-install-recommends \
  ca-certificates \
  git \
  unzip \
  upx-ucl \
  wget \
 && mkdir /build

WORKDIR /build

COPY . .

RUN go build -o packer-builder-arm

ENV PACKER_VERSION 1.10.0

RUN if [ "$(uname -m)" = "aarch64" ]; then PACKER_ARCH="arm64"; else PACKER_ARCH="amd64"; fi && \
  wget https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_${PACKER_ARCH}.zip -q -O /tmp/packer.zip && \
  unzip /tmp/packer.zip -d /bin && \
  chmod +x /bin/packer && \
  rm /tmp/packer.zip

# COMPRESS WITH UPX
RUN upx-ucl --lzma /build/packer-builder-arm /bin/packer

FROM public.ecr.aws/lts/ubuntu:jammy

# hadolint ignore=DL3008
RUN apt-get update -qq \
 && apt-get install -qqy --no-install-recommends \
  ca-certificates \
  dosfstools \
  fdisk \
  gdisk \
  kpartx \
  libarchive-tools \
  parted \
  psmisc \
  qemu-utils \
  sudo \
  xz-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /build/packer-builder-arm /bin/packer /bin/
COPY --from=qemu_binaries /binaries/* /usr/bin/

# Enable detailed logging
ENV PACKER_LOG=1

# Enable plugin / download caching for consecutive runs
ENV PACKER_PLUGIN_PATH="/bin:/build/.packer_plugins"
ENV PACKER_CACHE_DIR="/build/.packer_cache"

ENTRYPOINT ["/entrypoint.sh"]
