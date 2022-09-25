package utils

import (
	"crypto/sha256"
	"io"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

func ExecuteCommand(name string, args ...string) ([]byte, error) {
	zap.S().Debugf("%s %s", name, strings.Join(args, " "))
	return exec.Command(name, args...).CombinedOutput()
}

func ExecuteCommandStdin(stdin io.Reader, name string, args ...string) ([]byte, error) {
	zap.S().Debugf("%s %s", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)

	cmd.Stdin = stdin

	return cmd.CombinedOutput()
}

func MergeStrings(lines ...string) string {
	prunedLines := []string{}
	for _, line := range lines {
		if line != "" {
			prunedLines = append(prunedLines, line)
		}
	}

	return strings.Join(prunedLines, "\n")
}

func OrStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func Sum256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func ReadFile(path string) ([]byte, error) {
	if path == "" {
		return []byte{}, nil
	}

	if path == "-" {
		return io.ReadAll(os.Stdin)
	}

	return os.ReadFile(path)
}
