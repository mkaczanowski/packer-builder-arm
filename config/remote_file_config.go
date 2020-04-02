//go:generate mapstructure-to-hcl2 -type RemoteFileConfig
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

// RemoteFileConfig describes remote file(s) used to build the image.
// Internally go-getter is being used to fetch files, so you can refer to:
//     https://godoc.org/github.com/hashicorp/go-getter
type RemoteFileConfig struct {
	FileChecksum     string   `mapstructure:"file_checksum" required:"true"`
	FileChecksumURL  string   `mapstructure:"file_checksum_url"`
	FileChecksumType string   `mapstructure:"file_checksum_type"`
	FileUrls         []string `mapstructure:"file_urls"`
	FileUnarchiveCmd []string `mapstructure:"file_unarchive_cmd"`
	TargetPath       string   `mapstructure:"file_target_path"`
	TargetExtension  string   `mapstructure:"file_target_extension"`
}

// Prepare remote file config
func (c *RemoteFileConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	if len(c.FileUrls) == 0 {
		errs = append(
			errs,
			errors.New("one of file_url or file_urls must be specified"),
		)
		return
	}

	// prevent auto-decompress
	for i := range c.FileUrls {
		u, err := url.Parse(c.FileUrls[i])
		if err != nil {
			errs = append(errs, err)
		}

		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			errs = append(errs, err)
		}

		q.Add("archive", "false")
		u.RawQuery = q.Encode()

		c.FileUrls[i] = u.String()
	}

	c.FileChecksumType = strings.ToLower(c.FileChecksumType)
	c.TargetExtension = strings.ToLower(c.TargetExtension)

	if c.FileChecksumType == "none" {
		warnings = append(
			warnings,
			"a checksum type of 'none' was specified. Specyfying checksum is highly recommended",
		)
		return
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
		errs = append(errs, fmt.Errorf("a checksum must be specified"))
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
			errs = append(errs, fmt.Errorf("couldn't extract checksum from checksum file"))
		} else {
			c.FileChecksumType = cksum.Type
			c.FileChecksum = hex.EncodeToString(cksum.Value)
		}
	}

	return warnings, errs
}
