GOARCH := $$(go env GOARCH)

linux_amd64: export GOARCH=amd64
linux_amd64: linux

linux_arm64: export GOARCH=arm64
linux_arm64: linux

linux_arm: export GOARCH=arm
linux_arm: linux

linux_i386: export GOARCH=386
linux_i386: linux

linux: export GOOS=linux
linux:
	go build -o bin/helm-manager-$$GOOS-$$GOARCH main.go

darwin_amd64: export GOARCH=amd64
darwin_amd64: darwin

darwin_arm64: export GOARCH=arm64
darwin_arm64: darwin

darwin: export GOOS=darwin
darwin:
	go build -o bin/helm-manager-$$GOOS-$$GOARCH main.go


windows_amd64: export GOARCH=amd64
windows_amd64: windows

windows_arm64: export GOARCH=arm64
windows_arm64: windows

windows_arm: export GOARCH=arm
windows_arm: windows

windows_i386: export GOARCH=386
windows_i386: windows

windows: export GOOS=windows
windows:
	go build -o bin/helm-manager-$$GOOS-$$GOARCH.exe main.go

all: clean
	$(MAKE) linux_amd64 
	$(MAKE) linux_arm64 
	$(MAKE) linux_arm 
	$(MAKE) linux_i386 
	$(MAKE) darwin_amd64 
	$(MAKE) darwin_arm64 
	$(MAKE) windows_amd64 
	$(MAKE) windows_arm64 
	$(MAKE) windows_arm 
	$(MAKE) windows_i386

clean:
	rm -rf bin
