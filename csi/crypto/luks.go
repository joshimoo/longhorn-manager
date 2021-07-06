package crypto

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func Open(volume, devicePath, passphrase string) (stdout, stderr []byte, err error) {
	return cryptSetupWithPassphrase(&passphrase,
		"luksOpen", devicePath, volume, "-d", "/dev/stdin")
}

func Close(volume string) (stdout, stderr []byte, err error) {
	return cryptSetup("luksClose", volume)
}

func Format(devicePath, passphrase string) (stdout, stderr []byte, err error) {
	return cryptSetupWithPassphrase(&passphrase,
		"-q", "luksFormat", "--type", "luks2", "--hash", "sha256",
		devicePath, "-d", "/dev/stdin")
}

func Status(volume string) (stdout, stderr []byte, err error) {
	return cryptSetup("status", volume)
}

func cryptSetup(args ...string) (stdout, stderr []byte, err error) {
	return cryptSetupWithPassphrase(nil, args...)
}

func cryptSetupWithPassphrase(passphrase *string, args ...string) (stdout, stderr []byte, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	nsenterArgs := []string{"-t 1", "--all", "cryptsetup"}
	nsenterArgs = append(nsenterArgs, args...)
	// cmd := exec.Command("cryptsetup", args...)
	cmd := exec.Command("nsenter", nsenterArgs...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if passphrase != nil {
		cmd.Stdin = strings.NewReader(*passphrase)
	}

	if err := cmd.Run(); err != nil {
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), fmt.Errorf("failed to run cryptsetup args: %v error: %v", args, err)
	}

	return stdoutBuf.Bytes(), nil, nil
}
