package api


type Server struct{}

var _ ServerInterface = &Server{}

func NewServer() *Server {
	return &Server{}
}
