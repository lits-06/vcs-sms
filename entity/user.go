package entity

type User struct {
	ID       string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email    string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(255);not null"`
}

func (User) TableName() string {
	return "users"
}
