source "arm" "rock64" {
  file_urls             = ["https://redirect.armbian.com/rock64/Jammy_current"]
  file_checksum         = "7685d9e128d46021f9d5fb8987e4bd91a6e9c85129a49b5c4d4868236e3cbc1b"
  file_checksum_type    = "sha256"
  file_target_extension = "xz"
  file_unarchive_cmd    = ["xz", "--decompress", "$ARCHIVE_PATH"]
  image_build_method    = "reuse"
  image_path            = "rock64-armbian-jammy.img"
  image_size            = "2.5G"
  image_type            = "dos"
  image_partitions {
    name         = "root"
    type         = "83"
    start_sector = "8192"
    filesystem   = "ext4"
    size         = "2.37"
    mountpoint   = "/"
  }
  image_chroot_env             = ["PATH=/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin"]
  qemu_binary_source_path      = "/usr/bin/qemu-aarch64-static"
  qemu_binary_destination_path = "/usr/bin/qemu-aarch64-static"
}

build {
  sources = ["source.arm.rock64"]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y docker.io"     
    ]
  }
}
