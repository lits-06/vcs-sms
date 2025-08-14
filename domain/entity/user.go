package entity

type User struct {
	ID       string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email    string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Username string  `gorm:"type:varchar(50);not null"`
	Password string  `gorm:"type:varchar(255);not null"`
	RoleID   string  `gorm:"type:uuid;not null"`
	Role     Role    `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Scopes   []Scope `gorm:"many2many:user_scopes;constraint:OnDelete:CASCADE"`
}

type Role struct {
	ID    string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name  string `gorm:"type:varchar(50);not null"`
	Users []User `gorm:"foreignKey:RoleID"`
}

type Scope struct {
	ID    string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name  string `gorm:"type:varchar(50);not null"`
	Users []User `gorm:"many2many:user_scopes;constraint:OnDelete:CASCADE"`
}

func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
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
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

func DefaultUserRole() *Role {
	return &Role{Name: RoleUser}
}

func DefaultUserScopes() []string {
	return []string{ServerScopeView, ServerScopeExport}
}
