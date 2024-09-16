package assets

import (
	"atlas/pkg/logging"
	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog"
)

var (
	// Config is the configuration for the server.
	Config = loadConfig()
)

type C struct {
	Server struct {
		Debug      bool `toml:"debug"`
		ListenPort int  `toml:"listen_port"`
	} `toml:"server"`
	Ipv6 struct {
		Subnet string `toml:"subnet"`
	} `toml:"ipv6"`
	Database struct {
		Organization string `toml:"organization"`
	} `toml:"database"`
}

func loadConfig() C {
	var data C

	if _, err := toml.DecodeFile("assets/config.toml", &data); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Failed to load configuration file")
	} else {
		logging.Logger.Info().Str("file", "assets/config.toml").Msg("Loaded configuration file")
	}

	if data.Server.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return data
}
