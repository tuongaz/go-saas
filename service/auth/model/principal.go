package model

const (
	RoleAdmin   Role = "ADMIN"
	RoleMember  Role = "MEMBER"
	RoleOwner   Role = "OWNER"
	RoleService Role = "SERVICE"

	AccountTypeUser    = "USER"
	AccountTypeService = "SERVICE"
)

type Role string

func (r Role) IsOwner() bool {
	return r == RoleOwner
}

type Principal struct {
	OrganisationID string
	AccountID      string
	AccountType    string
	Role           Role
}
