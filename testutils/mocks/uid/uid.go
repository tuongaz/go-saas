package uid

import "github.com/tuongaz/go-saas/pkg/uid"

var _ uid.Interface = &mockUID{}

type mockUID struct {
	id string
}

func MockUID(id string) {
	uid.SetDefaultUID(&mockUID{id: id})
}

func ResetUID() {
	uid.SetDefaultUID(uid.Default)
}

func (m *mockUID) Generate() string {
	return m.id
}
