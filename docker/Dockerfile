FROM multiarch/qemu-user-static:register AS register
FROM golang:buster AS builder

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

ENV PACKER_VERSION 1.5.5

RUN wget https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_amd64.zip -O /tmp/packer.zip && \
  unzip /tmp/packer.zip -d /bin && \
  rm /tmp/packer.zip

# COMPRESS WITH UPX
RUN upx-ucl /build/packer-builder-arm /bin/packer

FROM ubuntu:focal

# hadolint ignore=DL3008
RUN apt-get update -qq \
 && apt-get install -qqy --no-install-recommends \
  qemu-user-static \
  binfmt-support \
  dosfstools \
  gdisk \
  kpartx \
  libarchive-tools \
  sudo \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /build/packer-builder-arm /bin/packer /bin/
COPY --from=register /register /register
COPY --from=register /qemu-binfmt-conf.sh /qemu-binfmt-conf.sh

ENTRYPOINT ["/entrypoint.sh"]
