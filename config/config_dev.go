package config

const (
	DB_HOST     = "localhost"
	DB_PORT     = "5432"
	DB_USER     = "dev_user"
	DB_PASSWORD = "dev_password"
	DB_NAME     = "sms_db"
	DB_SSL_MODE = "disable"

	LOG_FILE        = "sms_dev.log"
	LOG_MAX_SIZE    = 1 // in MB
	LOG_MAX_BACKUPS = 10
	LOG_MAX_AGE     = 30 // in days
	LOG_LEVEL       = "debug"
	LOG_COMPRESSION = true
)
