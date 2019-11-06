package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/packer/template/interpolate"
)

// By default, Packer will symlink, download or copy image files to the Packer
// cache into a "`hash($file+url+$file+checksum).$file+target_extension`" file.
// Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
// file mode in order to perform a download.
//
// go-getter supports the following protocols:
//
// * Local files
// * Git
// * Mercurial
// * HTTP
// * Amazon S3
//
//
// \~&gt; On windows - when referencing a local iso - if packer is running
// without symlinking rights, the iso will be copied to the cache folder. Read
// [Symlinks in Windows 10
// !](https://blogs.windows.com/buildingapps/2016/12/02/symlinks-windows-10/)
// for more info.
//
// Examples:
// go-getter can guess the checksum type based on `file+checksum` len.
//
// ``` json
// {
//   "file+checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
//   "file+url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
// ``` json
// {
//   "file+checksum_type": "file",
//   "file+checksum": "ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum",
//   "file+url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
// ``` json
// {
//   "file+checksum_url": "./shasums.txt",
//   "file+url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
// ``` json
// {
//   "file+checksum_type": "sha256",
//   "file+checksum_url": "./shasums.txt",
//   "file+url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
type RemoteFileConfig struct {
	// The checksum for the ISO file or virtual hard drive file. The algorithm
	// to use when computing the checksum can be optionally specified with
	// `file+checksum_type`. When `file+checksum_type` is not set packer will
	// guess the checksumming type based on `file+checksum` length.
	// `file+checksum` can be also be a file or an URL, in which case
	// `file+checksum_type` must be set to `file`; the go-getter will download
	// it and use the first hash found.
	FileChecksum string `mapstructure:"file_checksum" required:"true"`
	// An URL to a checksum file containing a checksum for the ISO file. At
	// least one of `file+checksum` and `file+checksum_url` must be defined.
	// `file+checksum_url` will be ignored if `file+checksum` is non empty.
	FileChecksumURL string `mapstructure:"file_checksum_url"`
	// The algorithm to be used when computing the checksum of the file
	// specified in `file+checksum`. Currently, valid values are "", "none",
	// "md5", "sha1", "sha256", "sha512" or "file". Since the validity of ISO
	// and virtual disk files are typically crucial to a successful build,
	// Packer performs a check of any supplied media by default. While setting
	// "none" will cause Packer to skip this check, corruption of large files
	// such as ISOs and virtual hard drives can occur from time to time. As
	// such, skipping this check is not recommended. `file+checksum_type` must
	// be set to `file` when `file+checksum` is an url.
	FileChecksumType string `mapstructure:"file_checksum_type"`
	// Multiple URLs for the ISO to download. Packer will try these in order.
	// If anything goes wrong attempting to download or while downloading a
	// single URL, it will move on to the next. All URLs must point to the same
	// file (same checksum). By default this is empty and `file+url` is used.
	// Only one of `file+url` or `file+urls` can be specified.
	FileUrls         []string `mapstructure:"file_urls"`
	FileUnarchiveCmd []string `mapstructure:"file_unarchive_cmd"`
	// The path where the iso should be saved after download. By default will
	// go in the packer cache, with a hash of the original filename and
	// checksum as its name.
	TargetPath string `mapstructure:"file_target_path"`
	// The extension of the iso file after download. This defaults to `raw`.
	TargetExtension string `mapstructure:"file_target_extension"`
}

func (c *RemoteFileConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	if len(c.FileUrls) == 0 {
		errs = append(
			errs, errors.New("One of file_url or file_urls must be specified"))
		return
	}

	c.FileChecksumType = strings.ToLower(c.FileChecksumType)
	c.TargetExtension = strings.ToLower(c.TargetExtension)

	// Warnings
	if c.FileChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
		return warnings, errs
	}

	if c.FileChecksumURL != "" {
		if c.FileChecksum != "" {
			warnings = append(warnings, "You have provided both an "+
				"file_checksum and an file+checksum_url. Discarding the "+
				"file_checksum_url and using the checksum.")
		} else {
			// go-getter auto-parses checksum files
			c.FileChecksumType = "file"
			c.FileChecksum = c.FileChecksumURL
		}
	}

	if c.FileChecksum == "" {
		errs = append(errs, fmt.Errorf("A checksum must be specified"))
	}
	if c.FileChecksumType == "file" {
		u, err := url.Parse(c.FileUrls[0])
		if err != nil {
			errs = append(errs, fmt.Errorf("error parsing URL <%s>: %s",
				c.FileUrls[0], err))
		}
		wd, err := os.Getwd()
		if err != nil {
			log.Printf("get working directory: %v", err)
			// here we ignore the error in case the
			// working directory is not needed.
		}
		gc := getter.Client{
			Dst:     "no-op",
			Src:     u.String(),
			Pwd:     wd,
			Dir:     false,
			Getters: getter.Getters,
		}
		cksum, err := gc.ChecksumFromFile(c.FileChecksumURL, u)
		if cksum == nil || err != nil {
			errs = append(errs, fmt.Errorf("Couldn't extract checksum from checksum file"))
		} else {
			c.FileChecksumType = cksum.Type
			c.FileChecksum = hex.EncodeToString(cksum.Value)
		}
	}

	return warnings, errs
}
