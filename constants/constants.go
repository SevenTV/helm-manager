package constants

import "os"

var inTerm = func() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}()

var stdinUsed = func() bool {
	if fi, err := os.Stdin.Stat(); err != nil {
		return false
	} else if fi.Mode()&os.ModeNamedPipe != 0 {
		return true
	} else {
		return false
	}
}()

func InTerm() bool {
	return inTerm
}

func StdinUsed() bool {
	return stdinUsed
}
