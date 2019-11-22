package builder

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/mholt/archiver"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

// StepCompressArtifact generates rootfs archive if required
type StepCompressArtifact struct {
	ImageMountPointKey string

	exclusions map[string]bool
	state      multistep.StateBag
}

func (s *StepCompressArtifact) prepare(state multistep.StateBag) {
	config := state.Get("config").(*cfg.Config)
	exclusions := make(map[string]bool)

	for _, mount := range config.ImageConfig.ImageChrootMounts {
		exclusions[mount.DestinationPath] = true
	}

	s.state = state
	s.exclusions = exclusions
}

// isLocationExcluded checks if given location should be skipped for compression
// NOTE: this is naive approach that supports only top level directory check
func (s *StepCompressArtifact) isLocationExcluded(pth string) bool {
	_, ok := s.exclusions[pth]
	return ok
}

func (s *StepCompressArtifact) getSrcs() ([]string, error) {
	imageMountpoint := s.state.Get(s.ImageMountPointKey).(string)
	srcs := []string{}

	files, err := ioutil.ReadDir(imageMountpoint)
	if err != nil {
		return srcs, err
	}

	for _, file := range files {
		loc := filepath.Join("/", file.Name())
		if s.isLocationExcluded(loc) {
			continue
		}

		srcs = append(srcs, loc)
	}

	return srcs, nil
}

// Run the step
func (s *StepCompressArtifact) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.prepare(state)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*cfg.Config)

	imagePath := config.ImageConfig.ImagePath
	imageBase := filepath.Base(config.ImageConfig.ImagePath)
	imageExt := filepath.Ext(imagePath)
	imageMountpoint := state.Get(s.ImageMountPointKey).(string)

	if imageExt == ".img" {
		// no compression needed
		return multistep.ActionContinue
	}

	dir, err := ioutil.TempDir("", "compress-artifact")
	if err != nil {
		ui.Error(fmt.Sprintf("error while creating temporary dir: %v", err))
		return multistep.ActionHalt
	}
	defer os.RemoveAll(dir)

	var archiveErr error
	var dst string

	if imageExt == ".gz" {
		// create rootfs archive with tar
		dst = filepath.Join(dir, imageBase)
		cmd := []string{
			"tar",
			"-cpzf",
			dst,
		}

		for pth := range s.exclusions {
			cmd = append(cmd, fmt.Sprintf("--exclude=%s", filepath.Join(imageMountpoint, pth)))
		}

		cmd = append(cmd, "--one-file-system", "-C", imageMountpoint, ".")

		ui.Message(fmt.Sprintf("creating rootfs archive"))
		_, archiveErr = exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	} else {
		// create rootfs archive with archiver
		ui.Message("creating rootfs archive with archiver")
		dst = filepath.Join(dir, imageBase)

		srcs, err := s.getSrcs()
		if err != nil {
			ui.Error(fmt.Sprintf("error while filtering source files: %v", err))
			return multistep.ActionHalt
		}

		archiveErr = archiver.Archive(srcs, dst)
	}

	if archiveErr != nil {
		ui.Error(fmt.Sprintf("error while creating rootfs archive: %v", archiveErr))
		return multistep.ActionHalt
	}

	if _, err := exec.Command("mv", dst, imagePath).CombinedOutput(); err != nil {
		ui.Error(fmt.Sprintf("error while moving archive: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepCompressArtifact) Cleanup(state multistep.StateBag) {}
