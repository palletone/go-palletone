package mediatorplugin

import "gopkg.in/urfave/cli.v1"

var (
	StaleProductionFlag = cli.BoolFlag{
		Name:  "enable-stale-production",
		Usage: "Enable Verified Unit production, even if the chain is stale.",
	}
)

// config data for mediator plugin
type Config struct {
	EnableStaleProduction bool	// Enable Verified Unit production, even if the chain is stale.
//	RequiredParticipation float32	// Percent of mediators (0-99) that must be participating in order to produce
//	PrivateKey map[string]string	//	Tuple of [PublicKey, WIF private key]
}

// mediator plugin default config
var DefaultConfig = Config{
	EnableStaleProduction:	false,
//	PrivateKey:				map[string]string{"":""},
}

func SetMediatorPluginConfig(ctx *cli.Context, cfg *Config)  {
	switch  {
	case ctx.GlobalIsSet(StaleProductionFlag.Name):
		cfg.EnableStaleProduction = ctx.GlobalBool(StaleProductionFlag.Name)
	}
}
