package configure

/*
import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/naoina/toml"
)

//Load specify toml file
func loadConfig(file string, cfg *Config) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(cfg); err != nil {
		return err
	}
	return nil
}

//Load palletone.toml file, if file not exist, load default config
// func LoadDefaultConfig() (*Config, error) {
// 	file, _ := exec.LookPath(os.Args[0])
// 	path := filepath.Dir(file)
// 	configFile := filepath.Join(path, "palletone.toml")
// 	log.Println("TOML config file path:" + configFile)
// 	cfg := DefaultConfig
// 	err := LoadConfig(configFile, cfg)

// 	return cfg, err
// }

//Load palletone.toml file to over write default config value
// func LoadConfig2DefaultValue() error {
// 	cfg, err := LoadDefaultConfig()
// 	if err != nil {
// 		return err
// 	}
// 	DefaultConfig.Log = cfg.Log
// 	DefaultConfig.Dag = cfg.Dag
// 	DefaultConfig.Consensus = cfg.Consensus
// 	DefaultConfig.Node = cfg.Node
// 	return nil
// }

func LoadConfigFromFile(tomlFilePath string) (*Config, error) {
	if tomlFilePath == "" {
		file, _ := exec.LookPath(os.Args[0])
		path := filepath.Dir(file)
		tomlFilePath = filepath.Join(path, "palletone.toml")
		log.Println("TOML config file path: " + tomlFilePath + " ,start to load...")
	}
	cfg := DefaultConfig
	err := loadConfig(tomlFilePath, cfg)
	// log.Println(strings.Join(cfg.Log.OutputPaths, ","))
	// log.Println(err)
	if err != nil {
		log.Fatal(err)
	}
	return cfg, err
}
*/
