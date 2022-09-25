package utils

import (
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/v2/constants"
	"github.com/seventv/helm-manager/v2/logger"
	"go.uber.org/zap"
)

type LoaderOptions struct {
	FetchingText string
	SuccessText  string
	FailureText  string
}

func Loader(options LoaderOptions) func(success bool) {
	downloading := make(chan bool)
	finished := make(chan struct{})
	go func() {
		defer close(downloading)
		defer close(finished)

		if constants.InTerm() {
			t := time.NewTicker(200 * time.Millisecond)
			defer t.Stop()
			i := 0
			stages := []string{"\\", "|", "/", "-"}
			for {
				select {
				case <-t.C:
					zap.S().Infof("%s [%s]\r", color.YellowString(options.FetchingText), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-downloading:
					if success {
						zap.S().Infof("%s %s", color.GreenString("✓"), options.SuccessText)
					} else {
						zap.S().Infof("%s %s", color.RedString("✗"), options.FailureText)
					}
					return
				}
			}
		} else {
			logger.Info(options.FetchingText)
			if <-downloading {
				logger.Info(options.SuccessText)
			} else {
				logger.Info(options.FailureText)
			}
		}
	}()

	once := sync.Once{}
	return func(success bool) {
		once.Do(func() {
			downloading <- success
			<-finished
		})
	}
}
