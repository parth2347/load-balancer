package load_balancer

type Server struct {
	Host string
	Port string
	Name string
	Scheme string
	Connections int
}

func (server Server) Url() string {
	return server.Scheme + "://" + server.Host + ":" + server.Port
}