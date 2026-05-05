package main

import (
	"os"
	"time"

	"github.com/Voxrame/panorama/internal/config"
	"github.com/Voxrame/panorama/internal/server"
	"github.com/Voxrame/panorama/static"
	"github.com/alexflint/go-arg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RunArgs struct{}

var args struct {
	ConfigPath string   `arg:"-c,--config" default:"config.toml"`
	Run        *RunArgs `arg:"subcommand:run"`
}

func main() {
	baseLogger:= createLogger()
	defer baseLogger.Sync()

	log := baseLogger.Sugar()

	arg.MustParse(&args)

	config, err := config.LoadConfig(args.ConfigPath)
	if err != nil {
		log.Errorw("Unable to load config", "error", err)
		os.Exit(1)
	}

	switch {
	case args.Run != nil:
		err = run(log, config)

	default:
		log.Warnw("Command not specified, proceeding with run")

		err = run(log, config)
	}

	if err != nil {
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger, config config.Config) error {
	quit := make(chan bool)

	log.Infow("starting web server", "address", config.Web.ListenAddress)

	go func() {
		server.Serve(static.UI, &config)
		quit <- true
	}()

	<-quit

	return nil
}

func createLogger() *zap.Logger {
	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("15:04:05"))
	}

	levelEncoder := func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		var text string

		switch l {
		case zapcore.DebugLevel:
			text = "\033[36mDEBUG\033[0m"
		case zapcore.InfoLevel:
			text = "\033[32mINFO \033[0m"
		case zapcore.WarnLevel:
			text = "\033[33mWARN \033[0m"
		case zapcore.ErrorLevel:
			text = "\033[31mERROR\033[0m"
		case zapcore.FatalLevel:
			text = "\033[35mFATAL\033[0m"
		}

		enc.AppendString(text)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "T",
		LevelKey:         "L",
		NameKey:          "N",
		CallerKey:        "C",
		MessageKey:       "M",
		StacktraceKey:    "S",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      levelEncoder,
		EncodeTime:       timeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)

	return zap.New(core)
}
