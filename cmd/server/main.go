package main

import (
	"atlas/internal/assets"
	"atlas/internal/handlers"
	"atlas/pkg/logging"
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", assets.Config.Server.ListenPort))
	if err != nil {
		logging.Logger.Error().Err(err).Msg("An unexpected error has occurred")
		return
	}

	logging.Logger.Info().
		Str("address", listener.Addr().String()).
		Msg("Listening for connections")

	for {
		c, err := listener.Accept()
		if err != nil {
			logging.Logger.Error().Err(err).Msg("An unexpected error has occurred")
			return
		}

		go handlers.Handle(c)
	}
}
