FROM tonistiigi/binfmt:qemu-v6.0.0 AS binfmt
FROM golang:1.17.5-buster AS builder

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

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o packer-builder-arm

ENV PACKER_VERSION 1.7.2

RUN if [ "$(uname -m)" = "aarch64" ]; then PACKER_ARCH="arm64"; else PACKER_ARCH="amd64"; fi && \
  wget https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_${PACKER_ARCH}.zip -q -O /tmp/packer.zip && \
  unzip /tmp/packer.zip -d /bin && \
  rm /tmp/packer.zip

# COMPRESS WITH UPX
RUN upx-ucl --lzma /build/packer-builder-arm /bin/packer

FROM public.ecr.aws/lts/ubuntu:focal

# hadolint ignore=DL3008
RUN apt-get update -qq \
 && apt-get install -qqy --no-install-recommends \
  ca-certificates \
  dosfstools \
  gdisk \
  kpartx \
  parted \
  libarchive-tools \
  sudo \
  xz-utils \
  psmisc \
  qemu-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /build/packer-builder-arm /bin/packer /bin/
COPY --from=binfmt /usr/bin/ /usr/bin

ENTRYPOINT ["/entrypoint.sh"]
