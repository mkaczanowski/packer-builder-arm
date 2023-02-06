source "arm" "ubuntu" {
  file_checksum         = "cb27bfa198b00ac265d3a1d38a4229f39f56646b3789847b3b84c805798466ef"
  file_checksum_type    = "sha256"
  file_target_extension = "xz"
  file_unarchive_cmd    = ["xz", "-d", "$ARCHIVE_PATH"]
  file_urls             = ["https://rcn-ee.net/rootfs/ubuntu-armhf-bionic-minimal/2023-02-05/am335x-ubuntu-18.04.6-console-armhf-2023-02-05-4gb.img.xz"]
  image_build_method    = "resize"
  image_path            = "bbb-sdcard-ubuntu-18.04.6-console.img"
  image_size            = "6G"
  image_type            = "dos"
  image_partitions {
    filesystem   = "ext4"
    mountpoint   = "/"
    name         = "root"
    size         = "0"
    start_sector = "8192"
    type         = "83"
  }
  image_chroot_env             = ["PATH=/usr/local/bin:/usr/local/sbin:/usr/bin:/bin:/sbin:/usr/sbin"]
  qemu_binary_destination_path = "/usr/bin/qemu-arm-static"
  qemu_binary_source_path      = "/usr/bin/qemu-arm-static"
}


build {
  sources = ["source.arm.ubuntu"]

  provisioner "shell" {
    inline = [
      "rm -f /etc/resolv.conf",
      "echo 'nameserver 1.1.1.1' > /etc/resolv.conf",
      "echo 'nameserver 8.8.8.8' >> /etc/resolv.conf",
      "apt-get update",
      "apt upgrade --yes --option=Dpkg::Options::=--force-confdef",
      "apt-get --yes autoremove",
      "apt-get --yes clean"
    ]
  }

}
