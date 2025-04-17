package uid

import (
	"github.com/segmentio/ksuid"
)

var Default Interface = New()

type Interface interface {
	Generate() string
}

func New() *UID {
	return &UID{}
}

type UID struct {
	Interface
}

func (u *UID) Generate() string {
	return ksuid.New().String()
}

func ID() string {
	return Default.Generate()
}

func SetDefaultUID(uid Interface) {
	Default = uid
}
