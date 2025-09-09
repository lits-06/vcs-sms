package domain

type User struct {
	ID       string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email    string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Username string  `gorm:"type:varchar(50);not null"`
	Password string  `gorm:"type:varchar(255);not null"`
	Scopes   []Scope `gorm:"many2many:user_scopes;constraint:OnDelete:CASCADE"`
}

type Scope struct {
	ID    string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name  string `gorm:"type:varchar(50);not null"`
	Users []User `gorm:"many2many:user_scopes;constraint:OnDelete:CASCADE"`
}

func (User) TableName() string {
	return "users"
}

func (Scope) TableName() string {
	return "scopes"
}

const (
	ServerScopeCreate = "server:create"
	ServerScopeView   = "server:view"
	ServerScopeUpdate = "server:update"
	ServerScopeDelete = "server:delete"
	ServerScopeImport = "server:import"
	ServerScopeExport = "server:export"
	ServerScopeReport = "server:report"

	UserScopeUpdate = "user:update"
)

func DefaultScopes() []Scope {
	return []Scope{
		{Name: ServerScopeView},
	}
}

func IsValidScope(scope string) bool {
	switch scope {
	case ServerScopeCreate,
		ServerScopeView,
		ServerScopeUpdate,
		ServerScopeDelete,
		ServerScopeImport,
		ServerScopeExport,
		ServerScopeReport:
		return true
	default:
		return false
	}
}
