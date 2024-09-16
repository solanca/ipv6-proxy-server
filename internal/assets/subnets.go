package assets

import (
	"atlas/pkg/logging"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Subnets is the list of subnets for the server.
	Subnets = loadSubnets()
)

func loadSubnets() map[string]string {
	subnets := map[string]string{}

	files, err := filepath.Glob("assets/subnets/*.txt")
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Msg("Error while reading subnets")
		os.Exit(1)
	}

	for _, fileName := range files {
		fileName = strings.ToLower(fileName)

		logging.Logger.Info().
			Str("file", fileName).
			Msg("Loading subnet file")

		file, err := os.Open(fileName)
		if err != nil {
			logging.Logger.Error().
				Err(err).
				Str("file", fileName).
				Msg("Error while reading subnets")
			os.Exit(1)
		}

		var result string

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			result += fmt.Sprintf("%v\n", scanner.Text())
		}

		file.Close()

		data := regexp.MustCompile(`name=(.*)\nsubnet=(.*)\n`).FindStringSubmatch(result)

		if len(data) >= 3 {
			name := data[1]
			subnet := data[2]

			logging.Logger.Info().
				Str("name", name).
				Str("subnet", subnet).
				Msg("Loaded subnet")

			subnets[strings.Replace(name, ".txt", "", -1)] = subnet
		} else {
			logging.Logger.Error().
				Str("file", fileName).
				Msg("Error while reading subnets")
		}

	}

	return subnets
}
