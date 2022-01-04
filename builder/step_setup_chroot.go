package builder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

func sortMountpoints(chrootMounts []cfg.ChrootMount, reverse bool) []cfg.ChrootMount {
	mounts := make([]cfg.ChrootMount, len(chrootMounts))
	copy(mounts, chrootMounts)

	sort.Slice(mounts, func(i, j int) bool {
		if reverse {
			return mounts[i].DestinationPath > mounts[j].DestinationPath
		}
		return mounts[i].DestinationPath < mounts[j].DestinationPath
	})

	return mounts
}

func prepareCmd(chrootMount cfg.ChrootMount, mountpoint string) []string {
	cmd := []string{
		"mount",
	}

	if chrootMount.MountType == "bind" {
		cmd = append(cmd, "--bind")
	} else if chrootMount.MountType == "rbind" {
		cmd = append(cmd, "--rbind")
	} else {
		cmd = append(cmd, "-t", chrootMount.MountType)
	}

	return append(cmd, chrootMount.SourcePath, mountpoint)
}

func getMounts() (map[string]bool, error) {
	dat, err := ioutil.ReadFile("/etc/mtab")
	if err != nil {
		return nil, err
	}

	selected := make(map[string]bool)
	all := strings.Split(string(dat), "\n")

	for _, line := range all {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		selected[string(fields[1])] = true
	}

	return selected, nil
}

// renameAndCheck courtesy of https://stackoverflow.com/a/25940392/51016
func renameAndCheck(src, dst string) error {
    err := os.Link(src, dst)
    if err != nil {
        return err
    }
    return os.Remove(src)
}
 
// deepCompare courtesy of https://stackoverflow.com/a/30038571/51016
const chunkSize = 64000

func deepCompare(file1, file2 string) (bool, error) {
    // Check file size ...
    f1, err := os.Open(file1)
    if err != nil {
        return false, err
    }
    defer f1.Close()

    f2, err := os.Open(file2)
    if err != nil {
        return false, err
    }
    defer f2.Close()

    for {
        b1 := make([]byte, chunkSize)
        _, err1 := f1.Read(b1)

        b2 := make([]byte, chunkSize)
        _, err2 := f2.Read(b2)

        if err1 != nil || err2 != nil {
            if err1 == io.EOF && err2 == io.EOF {
                return true, nil
            } else if err1 == io.EOF || err2 == io.EOF {
                return false, nil
            } else {
				err := fmt.Errorf("file1: %w", err1)
				err = fmt.Errorf("file2: %w", err2)
				err = fmt.Errorf("error comparing files; %w", err)
                return false, err
            }
        }

        if !bytes.Equal(b1, b2) {
            return false, nil
        }
    }
}

// StepSetupChroot prepares chroot environment by mounting specific locations (/dev /proc etc.)
type StepSetupChroot struct {
	ImageMountPointKey string
}

// Run the step
func (s *StepSetupChroot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	chrootMounts := sortMountpoints(config.ImageConfig.ImageChrootMounts, false)

	for _, chrootMount := range chrootMounts {
		mountpoint := filepath.Join(imageMountpoint, chrootMount.DestinationPath)
		cmd := prepareCmd(chrootMount, mountpoint)

		if err := os.MkdirAll(mountpoint, 0755); err != nil {
			ui.Error(fmt.Sprintf("error creating mount directory: %s", err))
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf("mounting %s with: %s", chrootMount.SourcePath, cmd))
		out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			ui.Error(fmt.Sprintf("error while mounting %v: %s", err, out))
			return multistep.ActionHalt
		}
	}

	ui.Message("patching file system for the chroot execution")

	src := "/etc/resolv.conf"
	dst := filepath.Join(imageMountpoint, src)
	bak := dst + ".bak"

	// backup the /etc/resolv.conf if it exists
	err := renameAndCheck(dst, bak)
	if err == nil {
		ui.Message(fmt.Sprintf("backed up '%s' to '%s'", dst, bak))
	} else if errors.Is(err, fs.ErrNotExist) {
		ui.Message(fmt.Sprintf("'%s' does not exist: %v", src, err))
	} else if errors.Is(err, fs.ErrExist) {
		ui.Error(fmt.Sprintf("'%s' already exists: %v", dst, err))
	} else if errors.Is(err, fs.ErrPermission) {
		ui.Error(fmt.Sprintf("could not create '%s': %v", bak, err))
	}

	source, err := os.Open(src)
	if err != nil {
		ui.Error(fmt.Sprintf("error while opening source: %v: '%s'", err, src))
		return multistep.ActionHalt
	}
	defer source.Close()

	// Should we backup the /etc/resolv.conf if it exist and restore it before creating the final image?
	destination, err := os.Create(dst)
	if err != nil {
		ui.Error(fmt.Sprintf("error while creating destination: %v: '%s'", err, dst))
		return multistep.ActionHalt
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err == nil {
		ui.Message(fmt.Sprintf("copied file from '%s' to '%s'", src, dst))
	} else {
		ui.Error(fmt.Sprintf("error while copying: %v: from '%s' to '%s'", err, src, dst))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepSetupChroot) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	chrootMounts := sortMountpoints(config.ImageConfig.ImageChrootMounts, true)
	imageMountpoint := state.Get(s.ImageMountPointKey).(string)

	// kill anything that would prevent the umount to succeed (best effort)
	out, err := exec.Command("fuser", "-k", imageMountpoint).CombinedOutput()
	if err != nil {
		ui.Message(fmt.Sprintf("optional (please ignore) `fuser -k` failed with %v: %s", err, out))
	}

	// read mtab and umount previously mounted targets
	mounted, err := getMounts()
	if err != nil {
		ui.Error(fmt.Sprintf("unable to read mtab: %v", err))
	}

	for _, chrootMount := range chrootMounts {
		mountpoint := filepath.Join(imageMountpoint, chrootMount.DestinationPath)
		if _, ok := mounted[mountpoint]; !ok {
			continue
		}

		for i := 0; i < 3; i++ {
			out, err := exec.Command("umount", mountpoint).CombinedOutput()
			if err != nil {
				if i == 2 {
					ui.Error(fmt.Sprintf("error while unmounting %v: %s", err, out))
				} else {
					// try to kill again (best effort)
					out, err = exec.Command("fuser", "-k", imageMountpoint).CombinedOutput()
					if err != nil {
						ui.Error(fmt.Sprintf("optional `fuser -k` failed with %v: %s", err, out))
					}
				}
			} else {
				break
			}
		}
	}

	// restore backed up /etc/resolve.conf in chroot
	resolve := filepath.Join(imageMountpoint, "/etc/resolv.conf")
	result, err := deepCompare("/etc/resolve.conf", resolve)
	if err != nil {
		ui.Error(fmt.Sprintf("error comparing host and chroot resolve.conf: %v", err))
	}
	// restore the backup only if the resolve.conf is unchanged
	if result {
		bak := resolve + ".bak"

		err = os.Remove(resolve)
		if err != nil {
			ui.Error(fmt.Sprintf("could not remove file %v: %s", err, resolve))
		}
		err = renameAndCheck(bak, resolve)
		if err == nil {
			ui.Message(fmt.Sprintf("restored '%s'", resolve))
		} else if errors.Is(err, fs.ErrNotExist) {
			ui.Error(fmt.Sprintf("'%s' does not exist: %v", bak, err))
		} else if errors.Is(err, fs.ErrExist) {
			ui.Error(fmt.Sprintf("'%s' already exists: %v", resolve, err))
		} else if errors.Is(err, fs.ErrPermission) {
			ui.Error(fmt.Sprintf("could not restore '%s': %v", resolve, err))
		}
	}
}
