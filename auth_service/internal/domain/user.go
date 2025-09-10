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

func ScopesToStringSlice(scopes []Scope) []string {
	result := make([]string, len(scopes))
	for i, scope := range scopes {
		result[i] = scope.Name
	}
	return result
}
