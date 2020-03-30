# Packer builder ARM

[![Build Status][travis-badge]][travis]
[![GoDoc][godoc-badge]][godoc]
[![GoReportCard][report-badge]][report]

[travis-badge]: https://travis-ci.org/mkaczanowski/packer-builder-arm.svg?branch=master
[travis]: https://travis-ci.org/mkaczanowski/packer-builder-arm/
[godoc-badge]: https://godoc.org/github.com/mkaczanowski/packer-builder-arm?status.svg
[godoc]: https://godoc.org/github.com/mkaczanowski/packer-builder-arm
[report-badge]: https://goreportcard.com/badge/github.com/mkaczanowski/packer-builder-arm
[report]: https://goreportcard.com/report/github.com/mkaczanowski/packer-builder-arm


This plugin allows you to build or extend ARM system image. It operates in two modes:
* new - creates empty disk image and populates the rootfs on it
* reuse - uses already existing image as the base

Plugin mimics standard image creation process, such as:
* builing base empty image (dd)
* partitioning (sgdisk / sfdisk)
* filesystem creation (mkfs.type)
* partition mapping (losetup)
* filesystem mount (mount)
* populate rootfs (tar/unzip/xz etc)
* setup qemu + chroot
* customize installation within chroot

The virtualization works via [binfmt_misc](https://en.wikipedia.org/wiki/Binfmt_misc) kernel feature and qemu.

Since the setup varies a lot for different hardware types, the example configuration is available per "board". Currently the following boards are supported (feel free to add more):
* bananapi-r1 (Archlinux ARM)
* beaglebone-black (Archlinux ARM, Debian)
* jetson-nano (Ubuntu)
* odroid-u3 (Archlinux ARM)
* odroid-xu4 (Archlinux ARM, Ubuntu)
* parallella (Ubuntu)
* raspberry-pi (Archlinux ARM, Raspbian)
* wandboard (Archlinux ARM)

# Quick start
```
git clone https://github.com/mkaczanowski/packer-builder-arm
cd packer-builder-arm
go mod download
go build

sudo packer build boards/odroid-u3/archlinuxarm.json
```
## Run in Docker
This method is primarily for macOS users where is no native way to use qemu-user-static.
Dockerfile is modeled on [packer-builder-arm-image](https://github.com/solo-io/packer-builder-arm-image).
### Build docker image:
```
docker build -t packer-builder-arm -f docker/Dockerfile .
```
### Usage:
Register qemu-user-static to kernel:
```
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```
Run packer build:
```
docker run --rm --privileged -v /dev:/dev -v ${PWD}:/build packer-builder-arm build boards/raspberry-pi/raspbian.json
```

# Dependencies
* `sfdisk / sgdisk`
* `e2fsprogs`

# Configuration
Configuration is split into 3 parts:
* remote file config
* image config
* qemu config

## Remote file
Describes the remote file that is going to be used as base image or rootfs archive (depending on `image_build_method`)

```
"file_urls" : ["http://os.archlinuxarm.org/os/ArchLinuxARM-odroid-xu3-latest.tar.gz"],
"file_checksum_url": "http://hu.mirror.archlinuxarm.org/os/ArchLinuxARM-odroid-xu3-latest.tar.gz.md5",
"file_checksum_type": "md5",
"file_unarchive_cmd": ["bsdtar", "-xpf", "$ARCHIVE_PATH", "-C", "$MOUNTPOINT"],
"file_target_extension": "tar.gz",
```

The `file_unarchive_cmd` is optional and should be used if the standard golang archiver can't handle the archive format.

Raw images format (`.img` or `.iso`) can be used by defining the `file_target_extension` appropriately.

## Image config
The base image description (size, partitions, mountpoints etc).

```
"image_build_method": "new",
"image_path": "odroid-xu4.img",
"image_size": "2G",
"image_type": "dos",
"image_partitions": [
    {
        "name": "root",
        "type": "8300",
        "start_sector": "4096",
        "filesystem": "ext4",
        "size": "0",
        "mountpoint": "/"
    }
],
```

The plugin doesn't try to detect the image partitions because that varies a lot. Instead it solely depend on `image_partitions` specification, so you should set that even if you reuse the image (`method` = reuse).

## Qemu config
Anything qemu related:

```
"qemu_binary_source_path": "/usr/bin/qemu-arm-static",
"qemu_binary_destination_path": "/usr/bin/qemu-arm-static"
```

# Chroot provisioner
To execute command within chroot environment you should use chroot communicator:
```
"provisioners": [
 {
   "type": "shell",
   "inline": [
     "pacman-key --init",
     "pacman-key --populate archlinuxarm"
   ]
 }
]
```

# Flashing
To dump image on device you can use [custom postprocessor](https://github.com/mkaczanowski/packer-post-processor-flasher) (really wrapper around `dd` with some sanity checks):
```
"post-processors": [
 {
     "type": "flasher",
     "device": "/dev/sdX",
     "block_size": "4096",
     "interactive": true
 }
]   
```

# Other
## Generating rootfs archive
While image (`.img`) format is useful for most cases, you might want to use
rootfs for other purposes (ex. export to docker). This is how you can generate
rootfs archive instead of image:
```
"image_path": "odroid-xu4.img" # generates image
"image_path": "odroid-xu4.img.tar.gz" # generates rootfs archive
```

## Docker
With `artifice` plugin you can pass rootfs archive to docker plugins
```
"post-processors": [
    [{
        "type": "artifice",
        "files": ["rootfs.tar.gz"]
    },
    {
        "type": "docker-import",
        "repository": "mkaczanowski/archlinuxarm",
        "tag": "latest"
    }],
    ...
]
```

## CI/CD
This is the live example on how to use github actions to push image to docker image registry:
```
cat .github/workflows/archlinuxarm-armv7-docker.yml
```

## How is this plugin different from `solo-io/packer-builder-arm-image`
https://github.com/hashicorp/packer/pull/8462

# Examples
For more examples please see:
```
tree boards/
```

# Demo
[![asciicast](https://asciinema.org/a/7ad1nm2Q7DRFVlHpqAknPolNo.svg)](https://asciinema.org/a/7ad1nm2Q7DRFVlHpqAknPolNo)
