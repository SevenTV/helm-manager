package utils

import (
	"bytes"
	"os"
	"path"

	"github.com/seventv/helm-manager/logger"
	"github.com/seventv/helm-manager/types"
	"gopkg.in/yaml.v3"
)

func ReadManifest(cwd string) error {
	data, err := os.ReadFile(path.Join(cwd, "manifest.yaml"))
	if err != nil {
		return nil
	}

	if err := yaml.Unmarshal(data, types.GlobalManifest); err != nil {
		return err
	}

	types.GlobalManifest.Exists = true

	return nil
}

func WriteManifest(cwd string) {
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)

	err := enc.Encode(types.GlobalManifest)
	if err != nil {
		logger.Fatal("failed to marshal manifest")
	}

	err = os.WriteFile(path.Join(cwd, "manifest.yaml"), buf.Bytes(), 0644)
	if err != nil {
		logger.Fatal("failed to write manifest")
	}
}
