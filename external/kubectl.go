package external

import (
	"bytes"
	"strings"

	"github.com/seventv/helm-manager/v2/utils"
)

type _kubectl struct{}

var Kubectl = _kubectl{}

func (_kubectl) GetNamespaces() ([]string, error) {
	data, err := utils.ExecuteCommand("kubectl", "get", "ns", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), " "), nil
}

func (_kubectl) Deploy(values []byte, namespace string, useCreate bool, dryRun bool) ([]byte, error) {
	args := []string{
		"apply",
		"-f",
		"-",
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	if useCreate {
		args = append(args, "--create")
	}

	return utils.ExecuteCommandStdin(bytes.NewReader(values), "kubectl", args...)
}

func (_kubectl) Delete(values []byte, namespace string, dryRun bool) ([]byte, error) {
	args := []string{
		"delete",
		"-f",
		"-",
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	return utils.ExecuteCommandStdin(bytes.NewReader(values), "kubectl", args...)
}
