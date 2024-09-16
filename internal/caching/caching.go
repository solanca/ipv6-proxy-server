package caching

import (
	"atlas/internal/assets"
	"atlas/pkg/goccm"
	"atlas/pkg/logging"
	"atlas/pkg/surrealdb"
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"github.com/tidwall/gjson"
	"time"
)

// init is called when the package is loaded and starts the update ticker
func init() {
	Concurrent = make(map[string]goccm.ConcurrencyManager)

	go func() {
		update()
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			update()
		}
	}()
}

// update retrieves the proxies from the database and updates the cache
func update() {
	response, err := surrealdb.Database.Select("proxies")
	if err != nil {
		logging.Logger.Error().
			Str("error", err.Error()).
			Msg("Unable to retrieve proxies from database.")
		return
	}

	data, err := json.Marshal(response)
	if err != nil {
		logging.Logger.Error().
			Str("error", err.Error()).
			Msg("Unable to marshal proxies from database.")
		return
	}

	proxies := gjson.Parse(string(data))
	for _, row := range proxies.Array() {
		if row.Get("organization").String() != assets.Config.Database.Organization {
			continue
		}

		mu.Lock()
		Accounts.Set(row.Get("username").String(), row, cache.DefaultExpiration)
		if _, ok := Concurrent[row.Get("username").String()]; !ok {
			Concurrent[row.Get("username").String()] = goccm.New(int(row.Get("threads").Int()))
		}
		mu.Unlock()
	}
}
