package domain

type ImportResponse struct {
	SuccessCount   int      `json:"success_count"`
	FailureCount   int      `json:"failure_count"`
	SuccessServers []string `json:"success_servers"`
	FailureServers []string `json:"failure_servers"`
}
