# Helm-Manager

Helm-Manager is a very simple cli tool used to manage helm chart value files.
It can automatically sync newer versions of charts onto older value files.

Look at the [manifest.example.yaml](./manifest.example.yaml) file to understand the structure.

Running this command without a `manifest.yaml` will cause it to create one.

## Purpose

The purpose of this tool is so that you can easily update helm chart values and changes are kept in sync.
Currently as far as I know there is no alternative to this tool.

## Requirements

You need to have the `helm` cli tool installed and on the executable must be on the path.

## Building

### Linux

```bash
make <linux/linux_amd64/linux_arm/linux_i386>
```

### Windows

```bash
make <windows/windows_amd64/windows_arm/windows_i386>
```

### Darwin

```bash
make <darwin/darwin_amd64/darwin_arm64>
```

## Installing

```bash
go install github.com/seventv/helm-manager
```

## Usage

### Base command

```bash
usage: helm-manager <Command> [-h|--help] [--debug] [-d|--working-dir
                    "<value>"] [-m|--manifest-file "<value>"] [-v|--values-dir
                    "<value>"]

                    Manage Helm Charts and their values

Commands:

  update  Use the update subcommand is used to update the values files or the
           cluster

Arguments:

  -h  --help           Print help information
      --debug          Enable debug logging
  -d  --working-dir    The working directory to use. Default: .
  -m  --manifest-file  The manifest file to use. Default: manifest.yaml
  -v  --values-dir     The values directory to use. Default: values
```

### Update subcommand

```bash
usage: helm-manager update [-h|--help] [--dry-run] [-t|--generate-template]
                    [-o|--template-output-dir "<value>"] [-i|--ignore-charts
                    "<value>" [-i|--ignore-charts "<value>" ...]]
                    [-f|--force-charts "<value>" [-f|--force-charts "<value>"
                    ...]] [-w|--wait] [-a|--atomic] [--no-stop] [--debug]
                    [-d|--working-dir "<value>"] [-m|--manifest-file "<value>"]
                    [-v|--values-dir "<value>"]

                    The update subcommand is used to update the values files or
                    the cluster

Arguments:

  -h  --help                 Print help information
      --dry-run              Dry run the upgrade
  -t  --generate-template    Generate a template file for the upgrade
  -o  --template-output-dir  The directory to output the generated template
                             files to. Default: templates
  -i  --ignore-charts        The charts to ignore
  -f  --force-charts         The charts to force upgrade
  -w  --wait                 Wait for the upgrade to complete
  -a  --atomic               Rollback the upgrade if it fails
      --no-stop              Disable stopping on the first error
      --debug                Enable debug logging
  -d  --working-dir          The working directory to use. Default: .
  -m  --manifest-file        The manifest file to use. Default: manifest.yaml
  -v  --values-dir           The values directory to use. Default: values
```
