package server

type server struct{}

func NewServer() *server {
	return &server{}
}

func (s *server) Run() error {
	return nil
}
