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

package config

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"github.com/palletone/go-palletone/core/vmContractPub/config"
	"github.com/palletone/go-palletone/core/vmContractPub/flogging"
)

// Config the config wrapper structure
type Config struct {
}

var configLogger = flogging.MustGetLogger("config")

func init() {

}

// SetupTestLogging setup the logging during test execution
func SetupTestLogging() {
	viper.SetDefault("logging.level", logging.DEBUG)
	level, err := logging.LogLevel(viper.GetString("logging.level"))
	if err == nil {
		// No error, use the setting
		logging.SetLevel(level, "main")
		logging.SetLevel(level, "server")
		logging.SetLevel(level, "peer")
	} else {
		configLogger.Warningf("Log level not recognized '%s', defaulting to %s: %s", viper.GetString("logging.level"), logging.ERROR, err)
		logging.SetLevel(logging.ERROR, "main")
		logging.SetLevel(logging.ERROR, "server")
		logging.SetLevel(logging.ERROR, "peer")
	}
}

// SetupTestConfig setup the config during test execution
func SetupTestConfig() {
	flag.Parse()

	// Now set the configuration file
	viper.SetEnvPrefix("CORE")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigName("core") // name of config file (without extension)
	err := config.AddDevConfigPath(nil)
	if err != nil {
		panic(fmt.Errorf("Fatal error adding DevConfigPath: %s \n", err))
	}

	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	SetupTestLogging()

	// Set the number of maxprocs
	var numProcsDesired = viper.GetInt("peer.gomaxprocs")
	configLogger.Debugf("setting Number of procs to %d, was %d\n", numProcsDesired, runtime.GOMAXPROCS(numProcsDesired))


	//glh
	// Init the BCCSP
	//var bccspConfig *factory.FactoryOpts
	//err = viper.UnmarshalKey("peer.BCCSP", &bccspConfig)
	//if err != nil {
	//	bccspConfig = nil
	//}
	//
	//tmpKeyStore, err := ioutil.TempDir("/tmp", "msp-keystore")
	//if err != nil {
	//	panic(fmt.Errorf("Could not create temporary directory: %s\n", tmpKeyStore))
	//}
	//
	//msp.SetupBCCSPKeystoreConfig(bccspConfig, tmpKeyStore)
	//
	//err = factory.InitFactories(bccspConfig)
	//if err != nil {
	//	panic(fmt.Errorf("Could not initialize BCCSP Factories [%s]", err))
	//}
}
