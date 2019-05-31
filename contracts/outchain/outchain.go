package outchain

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"path/filepath"

	"github.com/naoina/toml"
	"github.com/palletone/btc-adaptor"
	"github.com/palletone/eth-adaptor"
	"github.com/palletone/go-palletone/common/log"
)

type Config struct {
	Ada
}

type Ada struct {
	Btc      BTC
	Eth      ETH
	CCInfoKV map[string]CCInfo
}

type BTC struct {
	NetID        int
	Host         string
	RPCUser      string
	RPCPasswd    string
	CertPath     string
	WalletPasswd string

	ChaincodeKeys map[string]string
	AddressKeys   map[string]string
}
type ETH struct {
	NetID  int
	Rawurl string

	ChaincodeKeys map[string]string
	AddressKeys   map[string]string
}
type CCInfo struct {
	CCName      string
	ChainCodeKV map[string][]byte
}

var DefaultConfig = Config{
	Ada{
		Btc: BTC{
			NetID:         1,
			Host:          "localhost:18332",
			RPCUser:       "test",
			RPCPasswd:     "123456",
			CertPath:      "/home/pallet/wallet/btc/btctest/rpc.cert",
			WalletPasswd:  "1",
			ChaincodeKeys: map[string]string{},
			AddressKeys:   map[string]string{},
		},
		Eth: ETH{
			NetID:         1,
			Rawurl:        "/home/pallet/data/eth/gethtest/geth.ipc",
			ChaincodeKeys: map[string]string{},
			AddressKeys:   map[string]string{},
		},
		CCInfoKV: map[string]CCInfo{
			"test": CCInfo{
				ChainCodeKV: map[string][]byte{
					"testk": []byte("testv"),
				},
			},
		},
	},
}

const (
	modName    = "OutChain"
	configPath = "./outchain.toml"
)

var (
	gloaded = false
	cfg     = makeDefaultConfig()
)

type OutChainMethod struct {
	Method string `json:"method"`
}

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

func init() {
	f, err := os.Open(configPath)
	if err != nil && os.IsNotExist(err) {
		saveConfigTest() // save default config
	} else {
		f.Close()
	}

	//load config
	GetConfigTest()
}

func makeDefaultConfig() Config {
	return DefaultConfig
}

func loadConfig(file string, cfg *Config) error {
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

func makeConfigFile(cfg *Config, configPath string) error {
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
	if !gloaded {
		if err := loadConfig(configPath, &cfg); err != nil {
			return err
		}
		gloaded = true
	}
	return nil
}

func saveConfigTest() error {
	err := makeConfigFile(&cfg, configPath)
	if err != nil {
		log.Error("makeConfigFile() failed !!!!!!")
		return err
	}
	return nil
}

func GetJuryBTCPrikeyTest(chaincode string) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}
	var btcAdaptor adaptorbtc.AdaptorBTC
	btcAdaptor.NetID = cfg.Ada.Btc.NetID
	if _, ok := cfg.Ada.Btc.ChaincodeKeys[chaincode]; ok {
		return cfg.Ada.Btc.ChaincodeKeys[chaincode], nil
	} else {
		prikey := btcAdaptor.NewPrivateKey()
		addr := btcAdaptor.GetAddress(prikey)
		cfg.Ada.Btc.ChaincodeKeys[chaincode] = prikey
		cfg.Ada.Btc.AddressKeys[addr] = prikey
		// todo save config
		saveConfigTest()
		return prikey, nil
	}
	return "", errors.New("No private key of this chaicode")
}

func getJuryBTCPubkeyTest(chaincode string) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}
	var btcAdaptor adaptorbtc.AdaptorBTC
	btcAdaptor.NetID = cfg.Ada.Btc.NetID
	if _, ok := cfg.Ada.Btc.ChaincodeKeys[chaincode]; ok {
		pubkey := btcAdaptor.GetPublicKey(cfg.Ada.Btc.ChaincodeKeys[chaincode])
		return pubkey, nil
	} else {
		prikey := btcAdaptor.NewPrivateKey()
		pubkey := btcAdaptor.GetPublicKey(prikey)
		addr := btcAdaptor.GetAddress(prikey)
		cfg.Ada.Btc.ChaincodeKeys[chaincode] = prikey
		cfg.Ada.Btc.AddressKeys[addr] = prikey
		// todo save config
		saveConfigTest()
		return pubkey, nil
	}
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

func GetJuryETHPrikeyTest(chaincode string) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}
	var ethAdaptor adaptoreth.AdaptorETH
	ethAdaptor.NetID = cfg.Ada.Eth.NetID
	if _, ok := cfg.Ada.Eth.ChaincodeKeys[chaincode]; ok {
		return cfg.Ada.Eth.ChaincodeKeys[chaincode], nil
	} else {
		prikey := ethAdaptor.NewPrivateKey()
		addr := ethAdaptor.GetAddress(prikey)
		cfg.Ada.Eth.ChaincodeKeys[chaincode] = prikey
		cfg.Ada.Eth.AddressKeys[addr] = prikey
		// todo save config
		saveConfigTest()
		return prikey, nil
	}
	return "", errors.New("No private key of this chaicode")
}

func getJuryETHAddressTest(chaincode string) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}
	var ethAdaptor adaptoreth.AdaptorETH
	if _, ok := cfg.Ada.Eth.ChaincodeKeys[chaincode]; ok {
		addr := ethAdaptor.GetAddress(cfg.Ada.Eth.ChaincodeKeys[chaincode])
		return addr, nil
	} else {
		prikey := ethAdaptor.NewPrivateKey()
		addr := ethAdaptor.GetAddress(prikey)
		cfg.Ada.Eth.ChaincodeKeys[chaincode] = prikey
		cfg.Ada.Eth.AddressKeys[addr] = prikey
		// todo save config
		saveConfigTest()
		return addr, nil
	}
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

func GetChainCodeValue(chaincode string, key string) ([]byte, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return []byte(""), err
	}
	if _, ok := cfg.Ada.CCInfoKV[chaincode]; ok {
		if _, ok := cfg.Ada.CCInfoKV[chaincode].ChainCodeKV[key]; ok {
			return cfg.Ada.CCInfoKV[chaincode].ChainCodeKV[key], nil
		} else {
			return []byte(""), err
		}
	} else {
		return []byte(""), err
	}
	return []byte(""), errors.New("No private key of this chaicode")
}

func PutChainCodeValue(chaincode string, key string, value []byte) error {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return err
	}
	if _, ok := cfg.Ada.CCInfoKV[chaincode]; ok {
		cfg.Ada.CCInfoKV[chaincode].ChainCodeKV[key] = value
	} else {
		cfg.Ada.CCInfoKV[chaincode] = CCInfo{CCName: chaincode,
			ChainCodeKV: map[string][]byte{key: value}}
	}
	// todo save config
	saveConfigTest()

	return nil
}
