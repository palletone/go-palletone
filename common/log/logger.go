package log

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/palletone/go-palletone/dag/dagconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const errorKey = "ZAPLOG_ERROR"

var Logger *zap.Logger

type Plogger struct {
	logger zap.Logger
}

// New returns a new logger with the given context.
// New is a convenient alias for Root().New
func New(ctx ...interface{}) *Plogger {
	if Logger == nil {
		InitLogger()
	}
	pl := new(Plogger)
	pl.logger = *Logger
	return pl
}
func (pl *Plogger) New(ctx ...interface{}) *Plogger {
	if pl != nil {
		return pl
	}
	if Logger == nil {
		InitLogger()
	}

	pl.logger = *Logger
	return pl
}
func (pl *Plogger) Trace(msg string, ctx ...interface{}) {
	Trace(msg, ctx...)
}

func (pl *Plogger) Debug(msg string, ctx ...interface{}) {
	Debug(msg, ctx...)
}
func (pl *Plogger) Info(msg string, ctx ...interface{}) {
	Info(msg, ctx...)
}
func (pl *Plogger) Warn(msg string, ctx ...interface{}) {
	Warn(msg, ctx...)
}
func (pl *Plogger) Error(msg string, ctx ...interface{}) {
	Error(msg, ctx...)
}
func (pl *Plogger) Crit(msg string, ctx ...interface{}) {
	Crit(msg, ctx...)
}

// init zap.logger
func InitLogger() {
	// log path
	path := dagconfig.DefaultConfig.LoggerPath
	// error path
	err_path := dagconfig.DefaultConfig.ErrPath
	// log level
	lvl := dagconfig.DefaultConfig.LoggerLvl
	// is debug?
	isDebug := dagconfig.DefaultConfig.IsDebug
	log.Println("=============================================")
	log.Println("------------", path, err_path, lvl, isDebug, "------------")
	log.Println("=============================================")
	initLogger(path, err_path, lvl, isDebug)
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
}
func initLogger(path, err_path, lvl string, isDebug bool) {
	var js string
	if isDebug {
		js = fmt.Sprintf(`{
      "level": "%s",
      "encoding": "json",
      "outputPaths": ["stdout"],
      "errorOutputPaths": ["stdout"]
      }`, lvl)
	} else {
		js = fmt.Sprintf(`{
      "level": "%s",
      "encoding": "json",
      "outputPaths": ["%s"],
      "errorOutputPaths": ["%s"]
      }`, lvl, path, path)
	}
	var cfg zap.Config
	if err := json.Unmarshal([]byte(js), &cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var err error
	Logger, err = cfg.Build()
	if err != nil {
		log.Fatal("init logger error: ", err)
	}
}

// Trace
func Trace(msg string, ctx ...interface{}) {
	if Logger == nil {
		log.Println("logger is nil.")
		InitLogger()
	} else {
		//log.Println("logger trace is  ok.")
		fileds := ctxTOfileds(ctx...)

		Logger.Info(msg, fileds...)
	}
}

// Debug
func Debug(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		// log.Println("logger debug is ok.")
		fileds := ctxTOfileds(ctx...)

		Logger.Debug(msg, fileds...)
	}
}

// Info
func Info(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		// log.Println("logger info is ok.")
		fileds := ctxTOfileds(ctx...)

		Logger.Info(msg, fileds...)
	}
}

// Warn
func Warn(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		// log.Println("logger warn is ok.")
		fileds := ctxTOfileds(ctx...)

		Logger.Warn(msg, fileds...)
	}
}

// Error
func Error(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		// log.Println("logger error is ok.")
		fileds := ctxTOfileds(ctx...)

		Logger.Error(msg, fileds...)
	}
}

// Crit
func Crit(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		// log.Println("logger Crit is ok.")
		fileds := ctxTOfileds(ctx...)

		Logger.Info(msg, fileds...)
	}
}

// ctx transfer to  fileds
func ctxTOfileds(ctx ...interface{}) []zap.Field {
	// ctx translate into zap.Filed
	normalctx := normalize(ctx)
	fileds := make([]zap.Field, 0)
	var prefix, suffix []interface{}
	for i, v := range normalctx {
		if i%2 == 0 {
			prefix = append(prefix, v)
		} else {
			suffix = append(suffix, v)
		}
	}

	for i := 0; i < len(prefix); i++ {
		fileds = append(fileds, zap.Any(prefix[i].(string), suffix[i]))
	}
	return fileds
}

// normalize
func normalize(ctx []interface{}) []interface{} {
	// if the caller passed a Ctx object, then expand it
	if len(ctx) == 1 {
		if ctxMap, ok := ctx[0].(Ctx); ok {
			ctx = ctxMap.toArray()
		}
	}

	// ctx needs to be even because it's a series of key/value pairs
	// no one wants to check for errors on logging functions,
	// so instead of erroring on bad input, we'll just make sure
	// that things are the right length and users can fix bugs
	// when they see the output looks wrong
	if len(ctx)%2 != 0 {
		ctx = append(ctx, nil, errorKey, "Normalized odd number of arguments by adding nil")
	}

	return ctx
}

// Lazy allows you to defer calculation of a logged value that is expensive
// to compute until it is certain that it must be evaluated with the given filters.
//
// Lazy may also be used in conjunction with a Logger's New() function
// to generate a child logger which always reports the current value of changing
// state.
//
// You may wrap any function which takes no arguments to Lazy. It may return any
// number of values of any type.
type Lazy struct {
	Fn interface{}
}

// Ctx is a map of key/value pairs to pass as context to a log function
// Use this only if you really need greater safety around the arguments you pass
// to the logging functions.
type Ctx map[string]interface{}

func (c Ctx) toArray() []interface{} {
	arr := make([]interface{}, len(c)*2)

	i := 0
	for k, v := range c {
		arr[i] = k
		arr[i+1] = v
		i += 2
	}

	return arr
}
