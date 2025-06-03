package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/trinodb/trino-go-client/trino"
	_ "github.com/trinodb/trino-go-client/trino"
	"github.com/webvitals-sh/webvitals-edge-funcs/api/go/_internal/configs"
	"go.uber.org/fx"
)

func GetTrinoDSN(cfg *configs.Config) string {
	timeout := 5 * time.Minute
	trinoCfg := trino.Config{
		ServerURI: fmt.Sprintf("https://%s:%s@%s:%s",
			cfg.Starburst.User,
			cfg.Starburst.Password,
			cfg.Starburst.Host,
			cfg.Starburst.Port,
		),
		Catalog:      cfg.Starburst.Catalog,
		Schema:       cfg.Starburst.Schema,
		QueryTimeout: &timeout,
	}

	dsn, err := trinoCfg.FormatDSN()
	if err != nil {
		log.Fatalf("failed to build DSN: %v", err)
	}

	return dsn
}

func NewTrinoDB(lc fx.Lifecycle, cfg *configs.Config) *sql.DB {
	dsn := GetTrinoDSN(cfg)
	db, err := sql.Open("trino", dsn)
	if err != nil {
		log.Fatalf("failed to connect to trino: %v", err)
	}

	return db
}
