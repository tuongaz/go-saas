package postgres

import (
	"fmt"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/tuongaz/go-saas/service/auth/store/persistence"
)

var _ persistence.Interface = (*SQL)(nil)

type SQL struct {
	conn *sqlx.DB
}

func (s *SQL) Connection() *sqlx.DB {
	return s.conn
}

func New(datasource string) (*SQL, func(), error) {
	db, err := sqlx.Connect("postgres", datasource)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &SQL{
			conn: db,
		}, func() {
			_ = db.Close()
		}, nil
}
