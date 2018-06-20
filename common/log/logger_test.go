package log

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	initLogger("out.log", "error.log", "DEBUG", false)
	s := []string{
		"Hello info",
		"Hello error",
		"Hello debug",
		"Hello fatal",
	}
	Logger.Info("info", zap.String("str", s[0]), zap.Bool("bool1", true))
	Logger.Error("info", zap.String("str", s[1]))
	Logger.Error("info", zap.String("str", s[2]))
	Logger.Error("info", zap.String("str", s[3]))
}
func TestTrace(t *testing.T) {
	initLogger("out.log", "error.log", "DEBUG", false)
	Trace("test trace ...")
}

//InitLogger 初始化Logger
func InitLog(url string, optins ...interface{}) {
	if optins == nil {
		optins = append(optins, 0)
	}

	switch optins[0].(int) {
	case 1:
		logger, err := zap.NewProduction()
		if err != nil {
			panic(err)
		} else {
			fmt.Println("success")
		}

		//defer logger.Sync() // flushes buffer, if any
		logger.Named("log_palletone")
		sugar := logger.Sugar()
		sugar.Infow("failed to fetch URL",
			// Structured context as loosely typed key-value pairs.
			"url", url,
			"attempt", 3,
			"backoff", time.Second,
		)
		sugar.Infof("Failed to fetch URL: %s", url)
		sugar.Infof("ahhhhh")

		sugar.Warn("warning", zap.Error(errors.New("warn0")))
		sugar.Fatal("fatal", zap.String("fatal", "val_fatal"))

	default:
		Logger, err1 := zap.NewProduction()
		if err1 != nil {
			panic(err1)
		} else {
			fmt.Println("success")
		}
		//defer Logger.Sync()
		Logger.Named("log_palletone")
		Logger.Info("failed to fetch URL",
			// Structured context as strongly typed Field values.
			zap.String("url", url),
			zap.Int("attempt", 3),
			zap.Duration("backoff", time.Second),
		)
		Logger.Warn("warning", zap.Error(errors.New("warn1")))
		//Logger.Fatal("fatal", zap.String("fatal", "val_fatal"))
		Logger.Error("error ", zap.Error(errors.New("hahah")))
	}

}
