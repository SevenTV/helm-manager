package logger

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/seventv/helm-manager/constants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type writer struct {
	out *uilive.Writer
}

var Out io.Writer
var previousLine = struct {
	Data       []byte
	WasRewrite bool
}{}

func LoggerRewrite() {
	if len(previousLine.Data) != 0 && previousLine.WasRewrite {
		_, _ = Out.Write(previousLine.Data)
		previousLine.WasRewrite = false
	}
}

func (w *writer) Write(msg []byte) (int, error) {
	defer w.out.Flush()

	if len(msg) > 2 && msg[len(msg)-2] == '\r' {
		msg[len(msg)-2] = '\n'
		msg = msg[:len(msg)-1]

		previousLine.Data = msg
		previousLine.WasRewrite = true

		return w.out.Write(msg)
	}

	previousLine.Data = nil
	previousLine.WasRewrite = false

	return w.out.Bypass().Write(msg)
}

func init() {
	setupLogger(false)
}

func setupLogger(debug bool) {
	cfg := zap.NewProductionConfig()

	cfg.Encoding = "console"
	cfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05,000")
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.EncoderConfig.StacktraceKey = ""
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	lvl := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if debug {
		lvl = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		cfg.EncoderConfig.CallerKey = ""
		cfg.EncoderConfig.LevelKey = ""
		cfg.EncoderConfig.TimeKey = ""
	}

	Out = color.Output
	if constants.InTerm() {
		uilive.Out = Out
		out := uilive.New()
		Out = &writer{out: out}
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		zapcore.AddSync(Out),
		lvl,
	))

	zap.ReplaceGlobals(logger)
}

func SetDebug(debug bool) {
	setupLogger(debug)
}

func Debug(args ...any) {
	zap.S().Debugf("%s %s", color.New(color.Bold, color.FgBlack).Sprint(": "), color.MagentaString(fmt.Sprint(args...)))
}

func Debugf(format string, args ...any) {
	zap.S().Debugf("%s %s", color.New(color.Bold, color.FgBlack).Sprint(": "), color.MagentaString(format, args...))
}

func Info(args ...any) {
	zap.S().Infof("%s %s", color.New(color.Bold, color.FgBlack).Sprint(">"), color.WhiteString(fmt.Sprint(args...)))
}

func Infof(format string, args ...any) {
	zap.S().Infof("%s %s", color.New(color.Bold, color.FgBlack).Sprint(">"), color.WhiteString(format, args...))
}

func Warn(args ...any) {
	zap.S().Warnf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("->"), color.YellowString(fmt.Sprint(args...)))
}

func Warnf(format string, args ...any) {
	zap.S().Warnf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("->"), color.YellowString(format, args...))
}

func Error(args ...any) {
	zap.S().Errorf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("=>"), color.RedString(fmt.Sprint(args...)))
}

func Errorf(format string, args ...any) {
	zap.S().Errorf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("=>"), color.RedString(format, args...))
}

func Fatal(args ...any) {
	zap.S().Fatalf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("=>"), color.RedString(fmt.Sprint(args...)))
}

func Fatalf(format string, args ...any) {
	zap.S().Fatalf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("=>"), color.RedString(format, args...))
}
