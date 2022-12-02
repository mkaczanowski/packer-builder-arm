FROM --platform=$TARGETPLATFORM tonistiigi/binfmt:qemu-v7.0.0-28 AS binfmt
FROM --platform=$TARGETPLATFORM golang:1.19-bullseye AS builder

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

ENV PACKER_VERSION 1.8.4

RUN if [ "$(uname -m)" = "aarch64" ]; then PACKER_ARCH="arm64"; else PACKER_ARCH="amd64"; fi && \
  wget https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_${PACKER_ARCH}.zip -q -O /tmp/packer.zip && \
  unzip /tmp/packer.zip -d /bin && \
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
  #qemu-utils \
  sudo \
  xz-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /build/packer-builder-arm /bin/packer /bin/
COPY --from=binfmt /usr/bin/ /usr/bin

# Enable detailed logging
ENV PACKER_LOG=1

# Enable plugin / download caching for consecutive runs
ENV PACKER_PLUGIN_PATH="/build/.packer_plugins"
ENV PACKER_CACHE_DIR="/build/.packer_cache"

ENTRYPOINT ["/entrypoint.sh"]
