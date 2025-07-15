package validator

import (
	"github.com/lits-06/vcs-sms/entity"
)

type Validator interface {
	ValidateServer(s *entity.Server) error
	ValidateField(key string, value interface{}) error
	ValidateServerId(id string) error
	ValidateServerName(name string) error
	ValidateServerStatus(status string) error
	ValidateServerIpv4(ipv4 string) error
}
