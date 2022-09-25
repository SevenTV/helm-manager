# Helm-Manager

Helm-Manager is a very simple cli tool used to manage helm chart value files.
It can automatically sync newer versions of charts onto older value files.

Look at the [manifest.example.yaml](./manifest.example.yaml) file to understand the structure.

## Purpose

The purpose of this tool is so that you can easily update helm chart values and changes are kept in sync.
Currently as far as I know there is no alternative to this tool.

Essentially, this tool is designed to maintain `package.json` like manifest which can be used to easily upgrade the cluster's helm charts.

## Requirements

You need to have the `helm` and `kubectl` cli tools installed and on the executable must be on the path.

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
go install github.com/seventv/helm-manager/v2
```
