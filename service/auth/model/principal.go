package model

const (
	RoleOwner Role = "OWNER"
)

type Role string

func (r Role) IsOwner() bool {
	return r == RoleOwner
}

type Principal struct {
	OrganisationID string
	AccountID      string
	Role           Role
}
