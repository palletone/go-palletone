package nodeconfig

import (
	"path/filepath"

	"github.com/palletone/go-palletone/core/accounts/keystore"
)

const (
	// datadirPrivateKey      = "nodekey"            // Path within the datadir to the node's private key
	datadirDefaultKeyStore = "keystore" // Path within the datadir to the keystore
	// datadirStaticNodes     = "static-nodes.json"  // Path within the datadir to the static node list
	// datadirTrustedNodes    = "trusted-nodes.json" // Path within the datadir to the trusted node list
	// datadirNodeDatabase    = "nodes"              // Path within the datadir to store the node infos
)

//Config for node
type Config struct {
	DataDir     string
	KeyStoreDir string
	IPCPath     string
}

var DefaultConfig = Config{
	DataDir:     "/data1",
	KeyStoreDir: "/data1/keystore",
	IPCPath:     "gptn.ipc",
}

// AccountConfig determines the settings for scrypt and keydirectory
func (c *Config) AccountConfig() (int, int, string, error) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	// if c.UseLightweightKDF {
	// 	scryptN = keystore.LightScryptN
	// 	scryptP = keystore.LightScryptP
	// }

	var (
		keydir string
		err    error
	)
	switch {
	case filepath.IsAbs(c.KeyStoreDir):
		keydir = c.KeyStoreDir
	case c.DataDir != "":
		if c.KeyStoreDir == "" {
			keydir = filepath.Join(c.DataDir, datadirDefaultKeyStore)
		} else {
			keydir, err = filepath.Abs(c.KeyStoreDir)
		}
	case c.KeyStoreDir != "":
		keydir, err = filepath.Abs(c.KeyStoreDir)
	}
	return scryptN, scryptP, keydir, err
}
