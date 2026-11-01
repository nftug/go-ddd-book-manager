package config

import (
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/samber/lo"
)

type Config struct {
	DatabaseURL    string   `env:"DATABASE_URL,required,notEmpty"`
	OIDCIssuer     string   `env:"OIDC_ISSUER,required,notEmpty"`
	OIDCAudience   string   `env:"OIDC_AUDIENCE,required,notEmpty"`
	HTTPAddr       string   `env:"HTTP_ADDR" envDefault:":8080"`
	AllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS"`
}

func Load() (Config, error) {
	config, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	config.AllowedOrigins = lo.FilterMap(
		config.AllowedOrigins,
		func(origin string, _ int) (string, bool) {
			origin = strings.TrimSpace(origin)
			return origin, origin != ""
		},
	)
	return config, nil
}
