package domain

import (
	"context"
	"fmt"

	"github.com/valpere/trytrago/domain/database/repository"

	"github.com/valpere/trytrago/domain/database/repository/mysql"
	"github.com/valpere/trytrago/domain/database/repository/postgres"
	"github.com/valpere/trytrago/domain/database/repository/sqlite"
)

// NewRepository creates a new database repository based on the driver type
func NewRepository(ctx context.Context, opts repository.Options) (repository.Repository, error) {
	switch opts.Driver {
	case "postgres":
		return postgres.NewRepository(ctx, opts)
	case "mysql":
		return mysql.NewRepository(ctx, opts)
	case "sqlite":
		return sqlite.NewRepository(ctx, opts)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", opts.Driver)
	}
}
