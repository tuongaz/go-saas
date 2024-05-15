package postgres

import (
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

func NewFromDB(db *sqlx.DB) *SQL {
	return &SQL{
		conn: db,
	}
}
