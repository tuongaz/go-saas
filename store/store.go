package store

import (
	"github.com/autopus/bootstrap/store/persistence"
)

var _ Interface = (*Impl)(nil)

type Interface interface {
	AuthInterface
}

type Impl struct {
	db persistence.Interface
}

func New(db persistence.Interface) *Impl {
	return &Impl{
		db: db,
	}
}
