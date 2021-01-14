## Ubuntu Example
#
# Demonstrates:
#
# - Build comments
# - Image resizing
# - Network access
#
# Usage:
#
#     sudo podman run --rm \
#                     --privileged \
#                     --volume /dev:/dev \
#                     --volume ${PWD}:/build \
#                     --env POLICY_NAME=bbb \
#                     mkaczanowski/packer-builder-arm:latest \
#                     build ubuntu.hcl
#

"builders" = {
  "file_checksum" = "4c41976f1f574a4786002482167fe6443d8c9f1bced86a174459ffeba1f5b780"

  "file_checksum_type" = "sha256"

  "file_target_extension" = "xz"

  "file_unarchive_cmd" = ["xz", "-d", "$ARCHIVE_PATH"]

  "file_urls" = ["https://rcn-ee.com/rootfs/2020-03-12/microsd/bone-ubuntu-18.04.4-console-armhf-2020-03-12-2gb.img.xz"]

  ## Image Resizing
  #
  # Required.
  # Alternative values:
  #
  #  - reuse: fetch the image (.img) file, mount it and that's all. The size setting (below) is ignored in reuse mode.
  #  - new:   populate a rootfs (tarball with root file system).
  #
  "image_build_method" = "resize"

  "image_chroot_env" = ["PATH=/usr/local/bin:/usr/local/sbin:/usr/bin:/bin:/sbin:/usr/sbin"]

  "image_partitions" = {
    "filesystem" = "ext4"

    "mountpoint" = "/"

    "name" = "root"

    ## Image Resizing
    #
    # Required.
    #
    "size" = "0"

    "start_sector" = "8192"

    "type" = "83"
  }

  "image_path" = "bbb-sdcard-ubuntu-18.04.4-console.img"

  "image_setup_extra" = []

  "image_size" = "4G"

  "image_type" = "dos"

  "qemu_binary_destination_path" = "/usr/bin/qemu-arm-static"

  "qemu_binary_source_path" = "/usr/bin/qemu-arm-static"

  "type" = "arm"
}

"provisioners" = {
  "inline" = [
    "rm -f /etc/resolv.conf",
    "echo 'nameserver 1.1.1.1' > /etc/resolv.conf",
    "echo 'nameserver 8.8.8.8' >> /etc/resolv.conf",
    "apt-get update", 
    "apt upgrade --yes --option=Dpkg::Options::=--force-confdef", 
    "apt-get --yes autoremove", 
    "apt-get --yes clean"
  ]

  "type" = "shell"
}

"variables" = {}
