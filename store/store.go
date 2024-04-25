package store

import (
	"github.com/autopus/bootstrap/store/persistence"
)

var _ Interface = (*Impl)(nil)

type Interface interface {
	AuthInterface
	GetPersistence() persistence.Interface
}

type Impl struct {
	db persistence.Interface
}

func New(db persistence.Interface) *Impl {
	return &Impl{
		db: db,
	}
}

func (i *Impl) GetPersistence() persistence.Interface {
	return i.db
}
