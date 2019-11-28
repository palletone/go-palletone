/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

// log is the palletone log system.

package log

import (
	"fmt"
	"github.com/palletone/go-palletone/common/files"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"strings"
)

const (
	RootBuild      = "build"
	RootCmd        = "cmd"
	RootCommon     = "common"
	RootConfigure  = "configure"
	RootCore       = "core"
	RootInternal   = "internal"
	RootPtnclient  = "ptnclient"
	RootPtnjson    = "ptnjson"
	RootStatistics = "statistics"
	RootVendor     = "vendor"
	RootWallet     = "wallet"

	errorKey  = "ZAPLOG_ERROR"
	LogStdout = "stdout"
	LogStderr = "stderr"
)

//var defaultLogModule = []string{RootBuild, RootCmd, RootCommon, RootConfigure, RootCore, RootInternal,
// RootPtnclient, RootPtnjson, RootStatistics, RootVendor, RootWallet}

var LogConfig = DefaultConfig
var Logger *zap.Logger
//var mux sync.RWMutex

// init zap.logger
func InitLogger() {
	for _, path := range LogConfig.OutputPaths {
		//if path == LogStdout {
		//	continue
		//}

		if err := files.MakeDirAndFile(path); err != nil {
			panic(err)
		}
	}

	initLogger()
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
}

func ConsoleInitLogger() {
	LogConfig.LoggerLvl = "FATAL"

	outputPaths := make([]string, 0)
	for _, path := range LogConfig.OutputPaths {
		if path == LogStdout {
			continue
		}
		outputPaths = append(outputPaths, path)
	}
	LogConfig.OutputPaths = outputPaths

	errorPaths := make([]string, 0)
	for _, path := range LogConfig.ErrorOutputPaths {
		if path == LogStderr {
			continue
		}
		errorPaths = append(errorPaths, path)
	}
	LogConfig.ErrorOutputPaths = errorPaths

	InitLogger()
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
}

// init logger.
func initLogger() {
	var cfg zap.Config
	cfg.OutputPaths = LogConfig.OutputPaths
	cfg.ErrorOutputPaths = LogConfig.ErrorOutputPaths
	var lvl zap.AtomicLevel
	lvl.UnmarshalText([]byte(LogConfig.LoggerLvl))
	cfg.Level = lvl
	cfg.Encoding = LogConfig.Encoding
	cfg.Development = LogConfig.Development
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//cfg.EncoderConfig.EncodeLevel=zapcore.LowercaseColorLevelEncoder
	l, err := cfg.Build()
	if err != nil {
		log.Fatal("init logger error: ", err)
	}
	// add openModule
	//if strings.Contains(LogConfig.OpenModule[0], ",") {
	//	arr := strings.Split(LogConfig.OpenModule[0], ",")
	//	LogConfig.OpenModule[0] = ""
	//	LogConfig.OpenModule = append(LogConfig.OpenModule, arr...)
	//	LogConfig.OpenModule = append(LogConfig.OpenModule, defaultLogModule...)
	//} else {
	//	if !(len(LogConfig.OpenModule) == 1 && LogConfig.OpenModule[0] == "all") {
	//		LogConfig.OpenModule = append(LogConfig.OpenModule, defaultLogModule...)
	//	}
	//}
	//l.SetOpenModule(LogConfig.OpenModule)
	l = l.WithOptions(zap.AddCallerSkip(1))
	if LogConfig.RotationMaxSize > 0 {
		includeStdout, filePath := getOutputPath(LogConfig.OutputPaths)
		rotateLogCore := func(core zapcore.Core) zapcore.Core {
			mylogger := &RotationLogger{Max1LogLength: int64(LogConfig.MaxLogMessageLength)}
			mylogger.Logger = &lumberjack.Logger{
				Filename:   filePath,
				MaxSize:    LogConfig.RotationMaxSize, // megabytes
				MaxBackups: 60,
				MaxAge:     LogConfig.RotationMaxAge, // days
			}

			w := zapcore.AddSync(mylogger)
			if includeStdout {
				stdout, _, _ := zap.Open("stdout")
				w = zap.CombineWriteSyncers(stdout, w)
			}
			encoder := zapcore.NewConsoleEncoder(cfg.EncoderConfig)
			core2 := zapcore.NewCore(
				encoder,
				w,
				cfg.Level,
			)
			return core2
		}
		l = l.WithOptions(zap.WrapCore(rotateLogCore))
	}
	Logger = l
}

func getOutputPath(paths []string) (bool, string) {
	includeStdout := false
	filePath := ""
	for _, path := range paths {
		if strings.ToLower(path) == "stdout" {
			includeStdout = true
		} else {
			filePath = path
		}
	}
	return includeStdout, filePath
}

// Trace
func Trace(msg string, ctx ...interface{}) {
	if Logger == nil {
		//log.Println("logger is nil.")
		InitLogger()
	}
	//log.Println("logger trace is  ok.")
	fileds := ctxTOfileds(ctx...)
	Logger.Debug(msg, fileds...)
}

// Debug
func Debug(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	fileds := ctxTOfileds(ctx...)
	Logger.Debug(msg, fileds...)
}
func Debugf(format string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Debug(fmt.Sprintf(format, ctx...))
}

type GetString func() string

func DebugDynamic(getStr GetString) {
	if Logger == nil {
		InitLogger()
	}
	if LogConfig.LoggerLvl == "DEBUG" {
		Logger.Debug(getStr())
	}
}
func InfoDynamic(getStr GetString) {
	if Logger == nil {
		InitLogger()
	}
	if LogConfig.LoggerLvl == "DEBUG" || LogConfig.LoggerLvl == "INFO" {
		Logger.Info(getStr())
	}
}

// Info
func Info(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	fileds := ctxTOfileds(ctx...)
	Logger.Info(msg, fileds...)
}

func Infof(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Info(fmt.Sprintf(msg, ctx...))
}

// Warn
func Warn(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	fileds := ctxTOfileds(ctx...)
	Logger.Warn(msg, fileds...)
}
func Warnf(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Warn(fmt.Sprintf(msg, ctx...))
}

// Error
func Error(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}

	fileds := ctxTOfileds(ctx...)
	Logger.Error(msg, fileds...)
}

func Errorf(format string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Error(fmt.Sprintf(format, ctx...))
}

// Crit
func Crit(msg string, ctx ...interface{}) {
	if Logger == nil {
		InitLogger()
	}
	fileds := ctxTOfileds(ctx...)
	Logger.Error(msg, fileds...)
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
		pr := prefix[i]
		if e, ok := pr.(error); ok {
			fileds = append(fileds, zap.Any(e.Error(), suffix[i]))
		} else {
			fileds = append(fileds, zap.Any(pr.(string), suffix[i]))
		}
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
// type Lazy struct {
// 	Fn interface{}
// }

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

//用于限制一条Log的大小
type RotationLogger struct {
	Max1LogLength int64
	Logger        *lumberjack.Logger
}

func (l *RotationLogger) Write(p []byte) (n int, err error) {
	writeLen := int64(len(p))
	p1 := p
	if writeLen > l.Max1LogLength {
		p1 = append(p[0:l.Max1LogLength], []byte("...\r\n")...)
	}
	return l.Logger.Write(p1)
}
