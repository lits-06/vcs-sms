package dto

type CreateServerRequest struct {
	Name string `json:"name" validate:"required" example:"web-server-01"`      // Server name
	Port int    `json:"port" validate:"required,min=1,max=65535" example:"80"` // Port number
	IPv4 string `json:"ipv4" validate:"required,ipv4" example:"192.168.1.100"` // IPv4 address
}

type QueryServerRequest struct {
	Name   string `json:"name,omitempty" validate:"omitempty" form:"name" example:"web-server"`                         // Filter by server name
	Status string `json:"status,omitempty" validate:"omitempty,oneof=ON OFF" form:"status" example:"ON" enums:"ON,OFF"` // Filter by server status
	IPv4   string `json:"ipv4,omitempty" validate:"omitempty,ipv4" form:"ipv4" example:"192.168.1.100"`                 // Filter by IPv4 address

	From int `json:"from,omitempty" form:"from" example:"0"` // Pagination offset
	To   int `json:"to,omitempty" form:"to" example:"10"`    // Pagination limit

	Sort  string `json:"sort,omitempty" validate:"omitempty,oneof=name status created_at updated_at" form:"sort" example:"name" enums:"name,status,created_at,updated_at"` // Sort field
	Order string `json:"order,omitempty" validate:"omitempty,oneof=asc desc" form:"order" example:"asc" enums:"asc,desc"`                                                  // Sort order
}

type UpdateServerRequest struct {
	ID   string `json:"id" example:"uuid-string"`                                          // Server ID
	Name string `json:"name,omitempty" validate:"omitempty" example:"updated-server-name"` // Server name
	IPv4 string `json:"ipv4,omitempty" validate:"omitempty,ipv4" example:"192.168.1.101"`  // IPv4 address
}
