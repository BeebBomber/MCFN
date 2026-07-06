package deploy

import (
	"os"
	"os/exec"
)

func RemoteInstall(target, flakePath string) error {
	cmd := exec.Command("nixos-rebuild", "switch", "--flake", flakePath, "--target-host", target, "--use-remote-sudo")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
