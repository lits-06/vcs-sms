package domain

type QueryServerResponse struct {
	Servers *[]Server `json:"servers"`
	Total   int       `json:"total"`
}

type ImportResponse struct {
	SuccessCount   int
	FailureCount   int
	SuccessServers []string // format: "ID:Name"
	FailureServers []string // format: "ID:Name - error message"
}
