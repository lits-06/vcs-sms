package domain

import "github.com/lits-06/vcs-sms/domain/entity"

type CreateServerRequest struct {
	ID     string       `json:"id" validate:"required"`
	Name   string       `json:"name" validate:"required"`
	IPv4   string       `json:"ipv4" validate:"required,ipv4"`
	Status ServerStatus `json:"status" validate:"omitempty,oneof=ON OFF"`
}

type QueryServerRequest struct {
	Filter     ServerFilter     `json:"filter"`
	Pagination ServerPagination `json:"pagination"`
	Sort       ServerSort       `json:"sort"`
}

type UpdateServerRequest struct {
	ID     string              `json:"id"`
	Name   string              `json:"name,omitempty" validate:"omitempty"`
	IPv4   string              `json:"ipv4,omitempty" validate:"omitempty,ipv4"`
	Status entity.ServerStatus `json:"status,omitempty" validate:"omitempty,oneof=ON OFF"`
}
