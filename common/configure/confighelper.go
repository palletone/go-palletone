package configure

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/naoina/toml"
)

//Load specify toml file
func LoadConfig(file string, cfg *Config) error {
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
func LoadDefaultConfig() (*Config, error) {
	file, _ := exec.LookPath(os.Args[0])
	path := filepath.Dir(file)
	configFile := filepath.Join(path, "palletone.toml")
	log.Println("TOML config file path:" + configFile)
	cfg := DefaultConfig
	err := LoadConfig(configFile, cfg)

	return cfg, err
}

//Load palletone.toml file to over write default config value
func LoadConfig2DefaultValue() error {
	cfg, err := LoadDefaultConfig()
	if err != nil {
		return err
	}
	DefaultConfig.Log = cfg.Log
	return nil
}
