# Packer builder ARM
This plugin allows you to build or extend ARM system image. It operates in two modes:
* new - creates empty disk image and populates the rootfs on it
* reuse - uses already existing image as the base

Plugin mimics standard image creation process, such as:
* builing base empty image (dd)
* partitioning (sgdisk / sfdisk)
* filesystem creation (mkfs.<type>)
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

# Examples
For more examples please see:
```
tree boards/
```

# Flashing
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
