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
	"runtime"
	"strings"
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

func TestCheckFilePath(t *testing.T) {
	str := []string{
		"log",
		"log1/log.log",
		"log1",
		"log2.log",
	}
	re := []bool{
		false,
		false,
		true,
		true,
	}
	for i, v := range str {
		if re[i] == checkFileIsExist(v) {
			fmt.Println("success.")
		} else {
			fmt.Println("failed.")
		}
	}
	fmt.Println(runtime.GOOS)
}
func TestMkdirPath(t *testing.T) {
	paths := []string{
		"log/log.log ",
		"log1",
		"log2/log2_1/log2_2/log.log",
		"log1/log1_1/log.log",
	}

	for _, p := range paths {
		if mkdirPath("", p) == nil {
			fmt.Println("ture", p)
		} else {
			t.Error("false", p, mkdirPath(p, ""))
		}
	}
}
func TestMakeDirAndFile(t *testing.T) {
	paths := []string{
		"log2/log2_1/log2_2/log.log",
		"log1/log1_1/log.log",
	}
	for _, p := range paths {
		if MakeDirAndFile(p) == nil {
			fmt.Println("ture", p)
		} else {
			t.Error("false", p)
		}
	}
}

// Mkdir the path of out.log、err.log ,if the path is not exist.
func mkdirPath(path1, path2 string) error {
	var paths, errpaths []string
	oos := runtime.GOOS
	switch oos {
	case "windows":
		paths = strings.Split(path1, `\\`)
		errpaths = strings.Split(path2, `\\`)
	case "linux", "darwin":
		paths = strings.Split(path1, `/`)
		errpaths = strings.Split(path2, `/`)
	default:
		return errors.New("not supported on this system.")

	}
	if len(paths) > 0 {
		var path string
		for i, p := range paths {
			if i == 0 {
				path = p
			} else if i > 0 && i < len(paths)-1 {
				switch oos {
				case "windows":
					path += (`\\` + p)

				case "linux", "darwin":
					path += (`/` + p)
				}
			} else {
				break
			}

			if err := MakeDirAndFile(path); err != nil {
				return err
			}

		}
	}
	if len(errpaths) > 0 {
		var path string
		for i, e := range errpaths {
			if i == 0 {
				path = e
			} else if i > 0 && i < len(errpaths)-1 {
				switch oos {
				case "windows":
					path += (`\\` + e)

				case "linux", "darwin":
					path += (`/` + e)
				}
			} else {
				break
			}

			if err := MakeDirAndFile(path); err != nil {
				return err
			}

		}
	}
	return nil
}
