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
package log

import (
	"errors"
	"fmt"
	"testing"
	"time"

	//"github.com/palletone/go-palletone/common/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	InitLogger()
	s := []string{
		"Hello info",
		"Hello error",
		"Hello debug",
		"Hello fatal",
	}
	Logger.Info("msg_info", zap.String("str", s[0]), zap.Bool("bool1", true))
	Logger.Error("msg_error", zap.String("str", s[1]), zap.Bool("bool2", true))
	Logger.Debug("msg_debug", zap.String("str", s[2]))
	// Logger.Fatal("msg_fatal", zap.String("str", s[3]))
}
func TestTrace(t *testing.T) {
	InitLogger()
	Trace("test trace ...")
}

//InitLogger 初始化Logger
func InitLog(url string, optins ...interface{}) {
	if optins == nil {
		optins = append(optins, 0)
		fmt.Println("len:", len(optins))
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
func TestNewExample(t *testing.T) {
	l := zap.NewExample()
	l.Error("info", zap.String("1", "1"))
	l.Error("debug", zap.String("2", "2"))
	l.Error("error", zap.String("3", "3"))
}

func TestLogError(t *testing.T) {
	err := errors.New("Save data error.")
	//log := NewTestLog()
	Error("test error", err)
	Error("test error2", "test222", err)
}
func TestLogDynamic(t *testing.T) {
	LogConfig.LoggerLvl = "DEBUG"
	i := 0
	j := 0
	DebugDynamic(func() string { i++; return "debug called" })
	InfoDynamic(func() string {
		j++
		return "info called"
	})
	assert.Equal(t, i, 1)
	assert.Equal(t, j, 1)
	LogConfig.LoggerLvl = "INFO"
	DebugDynamic(func() string { i++; return "debug called" })
	InfoDynamic(func() string {
		j++
		return "info called"
	})
	assert.Equal(t, i, 1)
	assert.Equal(t, j, 2)
}

// func TestMain(m *testing.M) {
// 	//setup()
// 	code := m.Run()

// 	shutdown()
// 	os.Exit(code)
// }
// func shutdown() {
// 	fmt.Println("remove unit test log files.")
// 	os.RemoveAll("log")
// }
