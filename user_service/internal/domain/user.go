package domain

import (
	"fmt"

	"github.com/lits-06/vcs-sms/pkg/constants"
	"gorm.io/gorm"
)

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

func DefaultScopes() []Scope {
	return []Scope{
		{Name: constants.ServerScopeView},
	}
}

func IsValidScope(scope string) bool {
	switch scope {
	case constants.ServerScopeCreate,
		constants.ServerScopeView,
		constants.ServerScopeUpdate,
		constants.ServerScopeDelete,
		constants.ServerScopeImport,
		constants.ServerScopeExport,
		constants.ServerScopeReport:
		return true
	default:
		return false
	}
}

func InitScopes(db *gorm.DB) error {
	scopes := []string{
		constants.ServerScopeCreate,
		constants.ServerScopeView,
		constants.ServerScopeUpdate,
		constants.ServerScopeDelete,
		constants.ServerScopeImport,
		constants.ServerScopeExport,
		constants.ServerScopeReport,
		constants.UserScopeUpdate,
		constants.AdminScopeAll,
	}

	for _, scopeName := range scopes {
		var scope Scope
		if err := db.Where("name = ?", scopeName).First(&scope).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&Scope{Name: scopeName}).Error; err != nil {
					return fmt.Errorf("failed to create scope %s: %w", scopeName, err)
				}
			} else {
				return fmt.Errorf("failed to query scope %s: %w", scopeName, err)
			}
		}
	}

	return nil
}
