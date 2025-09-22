package domain

type ImportResponse struct {
	SuccessCount   int
	FailureCount   int
	SuccessServers []string
	FailureServers []string
}
