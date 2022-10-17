#!/usr/bin/env bash

set -o errtrace -o nounset -o pipefail -o errexit

echo "==> Setting nameserver"
mv /etc/resolv.conf /etc/resolv.conf.bk
echo 'nameserver 8.8.8.8' > /etc/resolv.conf

echo "==> Installing dependencies"
apt-get update -y
apt-get install -y apt-transport-https ca-certificates gnupg lsb-release

echo "==> Installing Docker"
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=arm64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update -y
apt-get install -y docker-ce docker-ce-cli containerd.io

echo "==> Cleaning"
apt-get clean
rm -rf /var/lib/apt/lists/*
