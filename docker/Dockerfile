FROM golang:buster AS builder
RUN apt-get update -qq \
 && apt-get install -qqy git && \
 mkdir /build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o packer-builder-arm

FROM ubuntu:eoan

RUN apt-get update -qq \
 && DEBIAN_FRONTEND=noninteractive apt-get install -qqy \
  qemu-user-static \
  kpartx \
  unzip \
  wget \
  curl \
  sudo \
 && rm -rf /var/lib/apt/lists/*

ENV PACKER_VERSION 1.4.5

RUN wget https://releases.hashicorp.com/packer/1.4.5/packer_1.4.5_linux_amd64.zip -O /tmp/packer.zip && \
  unzip /tmp/packer.zip -d /bin && \
  rm /tmp/packer.zip
WORKDIR /build
COPY docker/entrypoint.sh /entrypoint.sh

COPY --from=builder /build/packer-builder-arm /bin/packer-builder-arm

ENTRYPOINT ["/entrypoint.sh"]
