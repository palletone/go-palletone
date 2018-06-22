// Copyright 2018 The go-palletone Authors
// This file is part of go-palletone.
//
// go-palletone is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-palletone is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-palletone. If not, see <http://www.gnu.org/licenses/>.

// log is the palletone log system.

package log

import (
	"log"
	"os"
	"path"

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
	fileds := ctxTOfileds(ctx...)
	Logger.Info(msg, fileds...)
}

func (pl *Plogger) Debug(msg string, ctx ...interface{}) {
	fileds := ctxTOfileds(ctx...)
	Logger.Debug(msg, fileds...)
}
func (pl *Plogger) Info(msg string, ctx ...interface{}) {
	fileds := ctxTOfileds(ctx...)
	Logger.Info(msg, fileds...)
}
func (pl *Plogger) Warn(msg string, ctx ...interface{}) {
	fileds := ctxTOfileds(ctx...)
	Logger.Warn(msg, fileds...)
}
func (pl *Plogger) Error(msg string, ctx ...interface{}) {
	fileds := ctxTOfileds(ctx...)
	Logger.Error(msg, fileds...)
}
func (pl *Plogger) Crit(msg string, ctx ...interface{}) {
	fileds := ctxTOfileds(ctx...)
	Logger.Error(msg, fileds...)
}

// init zap.logger
func InitLogger() {
	// log path
	path := DefaultConfig.OutputPaths
	// error path
	err_path := DefaultConfig.ErrorOutputPaths
	// log level
	// lvl := DefaultConfig.LoggerLvl
	//  encoding
	// encoding := DefaultConfig.Encoding
	// // is debug?
	// isDebug := DefaultConfig.IsDebug
	// if the config file is damaged or lost, then initialize the config if log system.
	if len(path) == 0 {
		path = []string{"log/out.log"}
	}
	if len(err_path) == 0 {
		err_path = []string{"log/err.log"}
	}
	// if lvl == "" {
	// 	lvl = "INFO"
	// }
	// if encoding == "" {
	// 	encoding = "console"
	// }
	// if err := mkdirPath(path, err_path); err != nil {
	// 	panic(err)
	// }
	for _, p := range path {

		if err := MakeDirAndFile(p); err != nil {
			panic(err)
		}
	}
	for _, ep := range err_path {
		if err := MakeDirAndFile(ep); err != nil {
			panic(err)
		}
	}
	initLogger()
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
}

// init logger.
func initLogger() {
	// var js string
	// if isDebug {
	// 	js = fmt.Sprintf(`{
	//   "level": "%s",
	//   "encoding": "%s",
	//   "outputPaths": ["stdout","%s"],
	//   "errorOutputPaths": ["stderr","%s"]
	//   }`, lvl, encoding, path, err_path)
	// } else {
	// 	js = fmt.Sprintf(`{
	//   "level": "%s",
	//   "encoding": "%s",
	//   "outputPaths": ["%s"],
	//   "errorOutputPaths": ["%s"]
	//   }`, lvl, encoding, path, err_path)
	// }
	var cfg zap.Config
	//log.Println("Zap config json:" + js)
	// if err := json.Unmarshal([]byte(js), &cfg); err != nil {
	// 	panic(err)
	// }
	cfg.OutputPaths = DefaultConfig.OutputPaths
	cfg.ErrorOutputPaths = DefaultConfig.ErrorOutputPaths
	var lvl zap.AtomicLevel
	lvl.UnmarshalText([]byte(DefaultConfig.LoggerLvl))
	cfg.Level = lvl
	cfg.Encoding = DefaultConfig.Encoding
	cfg.Development = DefaultConfig.Development
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	l, err := cfg.Build()
	if err != nil {
		log.Fatal("init logger error: ", err)
	}
	Logger = l.WithOptions(zap.AddCallerSkip(1))
}

// Trace
func Trace(msg string, ctx ...interface{}) {
	if Logger == nil {
		//log.Println("logger is nil.")
		InitLogger()
	} else {
		//log.Println("logger trace is  ok.")
		fileds := ctxTOfileds(ctx...)
		Logger.Debug(msg, fileds...)
	}
}

// Debug
func Debug(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		fileds := ctxTOfileds(ctx...)
		Logger.Debug(msg, fileds...)
	}
}

// Info
func Info(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		fileds := ctxTOfileds(ctx...)
		Logger.Info(msg, fileds...)
	}
}

// Warn
func Warn(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		fileds := ctxTOfileds(ctx...)
		Logger.Warn(msg, fileds...)
	}
}

// Error
func Error(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		fileds := ctxTOfileds(ctx...)
		Logger.Error(msg, fileds...)
	}
}

// Crit
func Crit(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	} else {
		fileds := ctxTOfileds(ctx...)
		Logger.Error(msg, fileds...)
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

//CheckFileIsExist 判断文件是否存在，存在返回true，不存在返回false
func checkFileIsExist(path string) bool {
	var exist = true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// Mkdir the path of out.log、err.log ,if the path is not exist.
func MakeDirAndFile(filePath string) error {
	if filePath == "stdout" || filePath == "stderr" {
		return nil
	}
	// log.Println("log file path:" + filePath)
	if !checkFileIsExist(filePath) {
		// log.Println("create folder and file:" + filePath)
		err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
		if err != nil {
			return err
		}
		_, err = os.Create(filePath)
		if err != nil {
			return err
		}

	}
	return nil
}
