package mediatorplugin

import "gopkg.in/urfave/cli.v1"

var (
	StaleProductionFlag = cli.BoolFlag{
		Name:  "enable-stale-production",
		Usage: "Enable Verified Unit production, even if the chain is stale.",
	}
)

// mediator 结构体 和具体的账户模型有关
type Mediator struct {
	Address string
	Passphrase string
}

// config data for mediator plugin
type Config struct {
	EnableStaleProduction bool	// Enable Verified Unit production, even if the chain is stale.
//	RequiredParticipation float32	// Percent of mediators (0-99) that must be participating in order to produce
	Mediators []Mediator
}

// mediator plugin default config
var DefaultConfig = Config{
	EnableStaleProduction:	false,
	Mediators: []Mediator{
		{"P1XXX","123"},
	},
}

func SetMediatorPluginConfig(ctx *cli.Context, cfg *Config)  {
	switch  {
	case ctx.GlobalIsSet(StaleProductionFlag.Name):
		cfg.EnableStaleProduction = ctx.GlobalBool(StaleProductionFlag.Name)
	//case :
	//
	}
}
