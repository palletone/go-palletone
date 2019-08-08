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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package flogging

import (
	"os"
	"sync"
)

const (
	//pkgLogID      = "flogging"
	//defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
	defaultLevel = "DEBUG"
)

var (
	//log        log.ILogger
	defaultOutput *os.File

	modules map[string]string // Holds the map of all modules and their respective log level
	//peerStartModules map[string]string

	lock sync.RWMutex
	//once sync.Once
)

func init() {
	//log = log.New(pkgLogID)
	Reset()
	//initgrpclog()
}

// Reset sets to logging to the defaults defined in this package.
func Reset() {
	modules = make(map[string]string)
	lock = sync.RWMutex{}

	defaultOutput = os.Stderr
	//InitBackend(SetFormat(defaultFormat), defaultOutput)
	//InitFromSpec("")
}

// SetFormat sets the logging format.
//func SetFormat(formatSpec string) logging.Formatter {
//	if formatSpec == "" {
//		formatSpec = defaultFormat
//	}
//	return logging.MustStringFormatter(formatSpec)
//}

// InitBackend sets up the logging backend based on
// the provided logging formatter and I/O writer.
//func InitBackend(formatter logging.Formatter, output io.Writer) {
//	backend := logging.NewLogBackend(output, "", 0)
//	backendFormatter := logging.NewBackendFormatter(backend, formatter)
//	logging.SetBackend(backendFormatter).SetLevel(defaultLevel, "")
//}

// DefaultLevel returns the fallback value for logs to use if parsing fails.
func DefaultLevel() string {
	return defaultLevel
}

// GetModuleLevel gets the current logging level for the specified module.
func GetModuleLevel(module string) string {
	// logging.GetLevel() returns the logging level for the module, if defined.
	// Otherwise, it returns the default logging level, as set by
	// `flogging/logging.go`.
	level := defaultLevel
	return level
}

// SetModuleLevel sets the logging level for the modules that match the supplied
// regular expression. Can be used to dynamically change the log level for the
// module.
//func SetModuleLevel(moduleRegExp string, level string) (string, error) {
//	return setModuleLevel(moduleRegExp, level, true, false)
//}

//func setModuleLevel(moduleRegExp string, level string, isRegExp bool, revert bool) (string, error) {
//	var re *regexp.Regexp
//	logLevel, err := logging.LogLevel(level)
//	if err != nil {
//		log.Warningf("Invalid logging level '%s' - ignored", level)
//	} else {
//		if !isRegExp || revert {
//			logging.SetLevel(logLevel, moduleRegExp)
//			log.Debugf("Module '%s' log enabled for log level '%s'", moduleRegExp, level)
//		} else {
//			re, err = regexp.Compile(moduleRegExp)
//			if err != nil {
//				log.Warningf("Invalid regular expression: %s", moduleRegExp)
//				return "", err
//			}
//			lock.Lock()
//			defer lock.Unlock()
//			for module := range modules {
//				if re.MatchString(module) {
//					logging.SetLevel(logging.Level(logLevel), module)
//					modules[module] = logLevel.String()
//					log.Debugf("Module '%s' log enabled for log level '%s'", module, logLevel)
//				}
//			}
//		}
//	}
//	return logLevel.String(), err
//}

// MustGetLogger is used in place of `logging.MustGetLogger` to allow us to
// store a map of all modules and submodules that have logs in the system.
//func MustGetLogger(module string) log.ILogger {
//	l := log.New(module)
//	lock.Lock()
//	defer lock.Unlock()
//	modules[module] = GetModuleLevel(module)
//	return l
//}

// InitFromSpec initializes the logging based on the supplied spec. It is
// exposed externally so that consumers of the flogging package may parse their
// own logging specification. The logging specification has the following form:
//		[<module>[,<module>...]=]<level>[:[<module>[,<module>...]=]<level>...]
//func InitFromSpec(spec string) string {
//	levelAll := defaultLevel
//	var err error
//
//	if spec != "" {
//		fields := strings.Split(spec, ":")
//		for _, field := range fields {
//			split := strings.Split(field, "=")
//			switch len(split) {
//			case 1:
//				if levelAll, err = logging.LogLevel(field); err != nil {
//					log.Warningf("Logging level '%s' not recognized, defaulting to '%s': %s", field, defaultLevel, err)
//					levelAll = defaultLevel // need to reset cause original value was overwritten
//				}
//			case 2:
//				// <module>[,<module>...]=<level>
//				levelSingle, err := logging.LogLevel(split[1])
//				if err != nil {
//					log.Warningf("Invalid logging level in '%s' ignored", field)
//					continue
//				}
//
//				if split[0] == "" {
//					log.Warningf("Invalid logging override specification '%s' ignored - no module specified", field)
//				} else {
//					modules := strings.Split(split[0], ",")
//					for _, module := range modules {
//						log.Debugf("Setting logging level for module '%s' to '%s'", module, levelSingle)
//						logging.SetLevel(levelSingle, module)
//					}
//				}
//			default:
//				log.Warningf("Invalid logging override '%s' ignored - missing ':'?", field)
//			}
//		}
//	}
//
//	logging.SetLevel(levelAll, "") // set the logging level for all modules
//
//	// iterate through modules to reload their level in the modules map based on
//	// the new default level
//	for k := range modules {
//		MustGetLogger(k)
//	}
//	// register flogging log in the modules map
//	MustGetLogger(pkgLogID)
//
//	return levelAll.String()
//}

// SetPeerStartupModulesMap saves the modules and their log levels.
// this function should only be called at the end of peer startup.
//func SetPeerStartupModulesMap() {
//	lock.Lock()
//	defer lock.Unlock()
//
//	once.Do(func() {
//		peerStartModules = make(map[string]string)
//		for k, v := range modules {
//			peerStartModules[k] = v
//		}
//	})
//}

// GetPeerStartupLevel returns the peer startup level for the specified module.
// It will return an empty string if the input parameter is empty or the module
// is not found
//func GetPeerStartupLevel(module string) string {
//	if module != "" {
//		if level, ok := peerStartModules[module]; ok {
//			return level
//		}
//	}
//
//	return ""
//}

// RevertToPeerStartupLevels reverts the log levels for all modules to the level
// defined at the end of peer startup.
//func RevertToPeerStartupLevels() error {
//	lock.RLock()
//	defer lock.RUnlock()
//	for key := range peerStartModules {
//		_, err := setModuleLevel(key, peerStartModules[key], false, true)
//		if err != nil {
//			return err
//		}
//	}
//	log.Info("Log levels reverted to the levels defined at the end of peer startup")
//	return nil
//}
