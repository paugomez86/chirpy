package apiutils

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/paugomez86/chirpy/internal/database"
)

type ApiUtils struct {
	Cfg       Config
	DbQueries *database.Queries
}

type Config struct {
	Platform             string
	JwtSecret            string
	PolkaKey             string
	fileserverHits       atomic.Int32
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	blacklist            []string
}

// Initialize config struct
func (api *ApiUtils) LoadConfig() {
	api.Cfg.Platform = os.Getenv("PLATFORM")
	api.Cfg.JwtSecret = os.Getenv("JST_SECRET")
	api.Cfg.PolkaKey = os.Getenv("POLKA_KEY")
	api.Cfg.AccessTokenDuration = time.Hour
	api.Cfg.RefreshTokenDuration = time.Hour * 24 * 60
	api.loadBlacklist()
}

// Loads the blacklisted words in config struct
func (api *ApiUtils) loadBlacklist() {
	api.Cfg.blacklist = []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
}
