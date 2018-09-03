package outchain

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"path/filepath"

	"github.com/palletone/go-palletone/common/p2p"

	"github.com/naoina/toml"
	"github.com/palletone/go-palletone/adaptor"
	//"github.com/palletone/go-palletone/adaptor/btc-adaptor"
	//"github.com/palletone/go-palletone/adaptor/eth-adaptor"

	"github.com/palletone/go-palletone/common/log"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/statistics/dashboard"
)

type OutChainMethod struct {
	Method string `json:"method"`
}

type ptnstatsConfig struct {
	URL string `toml:",omitempty"`
}
type FullConfig struct {
	Node      node.Config
	Ptnstats  ptnstatsConfig
	Dashboard dashboard.Config
	//	Consensus consensusconfig.Config
	MediatorPlugin mp.Config
	Log            *log.Config
	Dag            *dagconfig.Config
	P2P            p2p.Config
	Ada            adaptor.Config
}

const modName = "OutChain"

var gloaded = false
var cfg = makeDefaultConfig()

var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

func makeDefaultConfig() FullConfig {
	return FullConfig{
		Ada: adaptor.DefaultConfig,
	}
}

func loadConfig(file string, cfg *FullConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func makeConfigFile(cfg *FullConfig, configPath string) error {
	var (
		configFile *os.File = nil
		err        error    = nil
	)

	err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
	if err != nil {
		return err
	}

	configFile, err = os.Create(configPath)
	defer configFile.Close()
	if err != nil {
		return err
	}

	configToml, err := tomlSettings.Marshal(cfg)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	_, err = configFile.Write(configToml)
	if err != nil {
		return err
	}

	return nil
}

func GetConfigTest() error {
	configPath := "F:/work/src/github.com/palletone/go-palletone/cmd/gptn/palletone.toml"
	if !gloaded {
		if err := loadConfig(configPath, &cfg); err != nil {
			return err
		}
		gloaded = true
	}
	return nil
}

func saveConfigTest() error {
	configPath := "F:/work/src/github.com/palletone/go-palletone/cmd/gptn/palletone.toml"
	err1 := makeConfigFile(&cfg, configPath)
	if err1 != nil {
		log.Error("makeConfigFile() failed !!!!!!")
	}
	return nil
}

func GetJuryBTCPrikeyTest(chaincode string) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}
	//var btcAdaptor adaptorbtc.AdaptorBTC
	//btcAdaptor.NetID = cfg.Ada.Btc.NetID
	//if _, ok := cfg.Ada.Btc.ChaincodeKeys[chaincode]; ok {
	//	return cfg.Ada.Btc.ChaincodeKeys[chaincode], nil
	//} else {
	//	prikey := btcAdaptor.NewPrivateKey()
	//	addr := btcAdaptor.GetAddress(prikey)
	//	cfg.Ada.Btc.ChaincodeKeys[chaincode] = prikey
	//	cfg.Ada.Btc.AddressKeys[addr] = prikey
	//	// todo save config
	//	saveConfigTest()
	//	return prikey, nil
	//}
	return "", errors.New("No private key of this chaicode")
}

func getJuryBTCPubkeyTest(chaincode string) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}
	//var btcAdaptor adaptorbtc.AdaptorBTC
	//btcAdaptor.NetID = cfg.Ada.Btc.NetID
	//if _, ok := cfg.Ada.Btc.ChaincodeKeys[chaincode]; ok {
	//	pubkey := btcAdaptor.GetPublicKey(cfg.Ada.Btc.ChaincodeKeys[chaincode])
	//	return pubkey, nil
	//} else {
	//	prikey := btcAdaptor.NewPrivateKey()
	//	pubkey := btcAdaptor.GetPublicKey(prikey)
	//	addr := btcAdaptor.GetAddress(prikey)
	//	cfg.Ada.Btc.ChaincodeKeys[chaincode] = prikey
	//	cfg.Ada.Btc.AddressKeys[addr] = prikey
	//	// todo save config
	//	saveConfigTest()
	//	return pubkey, nil
	//}
	return "", errors.New("No private key of this chaicode")
}

func ClolletJuryBTCPubkeysTest(chaincode string) ([]string, error) {
	var pubkeys []string
	pubkey, err := getJuryBTCPubkeyTest(chaincode)
	if err == nil {
		pubkeys = append(pubkeys, pubkey)
	}
	return pubkeys, nil
}

func getJuryETHAddressTest(chaincode string) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}
	//var ethAdaptor adaptoreth.AdaptorETH
	//if _, ok := cfg.Ada.Eth.ChaincodeKeys[chaincode]; ok {
	//	addr := ethAdaptor.GetAddress(cfg.Ada.Eth.ChaincodeKeys[chaincode])
	//	return addr, nil
	//} else {
	//	prikey := ethAdaptor.NewPrivateKey()
	//	addr := ethAdaptor.GetAddress(prikey)
	//	cfg.Ada.Eth.ChaincodeKeys[chaincode] = prikey
	//	cfg.Ada.Eth.AddressKeys[addr] = prikey
	//	// todo save config
	//	saveConfigTest()
	//	return addr, nil
	//}
	return "", errors.New("No private key of this chaicode")
}

func ClolletJuryETHAddressesTest(chaincode string) ([]string, error) {
	var addresses []string
	address, err := getJuryETHAddressTest(chaincode)
	if err == nil {
		addresses = append(addresses, address)
	}
	return addresses, nil
}
