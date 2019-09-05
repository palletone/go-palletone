package outchain

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"unicode"

	"github.com/naoina/toml"
	"github.com/palletone/go-palletone/common/log"

	"github.com/palletone/adaptor"
)

type Config struct {
	Ada
}

type Ada struct {
	Btc        BTC
	Eth        ETH
	ChainKeyKV map[string]KeyInfo //chainName --- keyInfo
}

type BTC struct {
	NetID        int
	Host         string
	RPCUser      string
	RPCPasswd    string
	CertPath     string
	WalletPasswd string
}
type ETH struct {
	NetID      int
	Rawurl     string
	TxQueryUrl string
}

type KeyInfo struct { //information of private key
	ChaincodeKeys map[string][]byte //chaincode --- privateKey
	AddressKeys   map[string][]byte //address --- privateKey
}

type CCInfo struct {
	CCName      string
	ChainCodeKV map[string][]byte
}

var DefaultConfig = Config{
	Ada{
		Btc: BTC{
			NetID:        1,
			Host:         "localhost:18332",
			RPCUser:      "test",
			RPCPasswd:    "123456",
			CertPath:     "/home/pallet/wallet/btc/btctest/rpc.cert",
			WalletPasswd: "1",
		},
		Eth: ETH{
			NetID:      1,
			Rawurl:     "https://ropsten.infura.io/",
			TxQueryUrl: "https://api-ropsten.etherscan.io/api?apikey=VYSBPQ383RJXM7HBQVTIK5NGIG8ZYVV6T6",
		},
		ChainKeyKV: map[string]KeyInfo{
			"btc": KeyInfo{
				ChaincodeKeys: map[string][]byte{},
				AddressKeys:   map[string][]byte{},
			},
			"eth": KeyInfo{
				ChaincodeKeys: map[string][]byte{},
				AddressKeys:   map[string][]byte{},
			},
		},
	},
}

const (
	//modName    = "OutChain"
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
	if err != nil {
		return err
	}
	defer configFile.Close()

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

func getNewKey(params []byte, iadaptor adaptor.ICryptoCurrency) ([]byte, string, error) {
	var input adaptor.NewPrivateKeyInput
	err := json.Unmarshal(params, &input)
	if err != nil {
		return []byte{}, "", fmt.Errorf("NewPrivateKeyInput params error : %s", err.Error())
	}
	outputKey, err := iadaptor.NewPrivateKey(nil)
	if err != nil {
		log.Error("NewPrivateKey() failed !!!!!!")
		return []byte{}, "", err
	}
	outputPub, err := iadaptor.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: outputKey.PrivateKey})
	if err != nil {
		log.Error("NewPrivateKey() failed !!!!!!")
		return []byte{}, "", err
	}
	outputAddr, err := iadaptor.GetAddress(&adaptor.GetAddressInput{Key: outputPub.PublicKey})
	if err != nil {
		log.Error("GetAddress() failed !!!!!!")
		return []byte{}, "", err
	}
	return outputKey.PrivateKey, outputAddr.Address, nil
}

func GetJuryKeyInfo(chaincode, chainName string, params []byte, iadaptor adaptor.ICryptoCurrency) ([]byte, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return []byte{}, err
	}

	if _, ok := cfg.Ada.ChainKeyKV[chainName]; !ok {
		cfg.Ada.ChainKeyKV[chainName] = KeyInfo{
			ChaincodeKeys: map[string][]byte{},
			AddressKeys:   map[string][]byte{},
		}
	}
	if _, exist := cfg.Ada.ChainKeyKV[chainName].ChaincodeKeys[chaincode]; exist {
		return cfg.Ada.ChainKeyKV[chainName].ChaincodeKeys[chaincode], nil
	} else {
		priKey, addr, err := getNewKey(params, iadaptor)
		if err != nil {
			log.Errorf("getNewKey() failed  %s !!!!!!", err.Error())
			return []byte{}, err
		}
		cfg.Ada.ChainKeyKV[chainName].ChaincodeKeys[chaincode] = priKey
		cfg.Ada.ChainKeyKV[chainName].AddressKeys[addr] = priKey

		saveConfigTest() // todo save config

		return priKey, nil
	}
}

func GetJuryAddress(chaincode, chainName string, params []byte, iadaptor adaptor.ICryptoCurrency) (string, error) {
	err := GetConfigTest()
	if err != nil {
		log.Error("loadconfig() failed !!!!!!")
		return "", err
	}

	if _, ok := cfg.Ada.ChainKeyKV[chainName]; ok {
		for addr := range cfg.Ada.ChainKeyKV[chainName].AddressKeys {
			return addr, nil
		}
		return "", err
	} else {
		priKey, addr, err := getNewKey(params, iadaptor)
		if err != nil {
			log.Error("getNewKey() failed !!!!!!")
			return "", err
		}
		cfg.Ada.ChainKeyKV[chainName].ChaincodeKeys[chaincode] = priKey
		cfg.Ada.ChainKeyKV[chainName].AddressKeys[addr] = priKey
		// todo save config
		saveConfigTest()
		return addr, nil
	}
}
