# This script allows users to go from a "blank" Debian 10 or Ubuntu 20.04 installation to a bare metal environment that cranks out device images with ease.
# Built for paranoid types who don't feel good about binary Docker images.  
# Used in https://github.com/faddat/imager

#!/bin/bash

# INSTALL GO
wget --progress=bar:force:noscroll https://golang.org/dl/go1.15.2.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.15.2.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.profile


# INSTALL DEPENDENCIES
DEBIAN_FRONTEND="noninteractive" apt install -y ca-certificates git unzip wget qemu-user-static build-essential qemu-user-static ca-certificates dosfstools gdisk kpartx parted libarchive-tools sudo xz-utils psmisc
git clone https://github.com/hashicorp/packer --branch v1.5.5
cd packer
go mod vendor
go get .
go install
cd ..

# INSTALL PACKER-BUILDER-ARM
git clone https://github.com//mkaczanowski/packer-builder-arm/
cd packer-builder-arm
go mod download
go build
sudo cp packer-builder-arm /usr/local/bin
cd ..


# INSTALL PISHRINK
wget https://raw.githubusercontent.com/Drewsif/PiShrink/master/pishrink.sh
chmod +x pishrink.sh
sudo mv pishrink.sh /usr/local/bin


echo "you're really going to want to source your bash profile like:"
echo "source ~/.profile"
echo "otherwise, Go won't work properly and you won't have a good time :)"
echo "have a good time!"
echo "example:"
echo "packer build boards/raspberry-pi-4/archlinuxarm.json"
