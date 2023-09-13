package server

type Config = func(*Server) error

func WithAddr(addr string) Config {
	return func(s *Server) error {
		s.Addr = addr

		return nil
	}
}
 var _ = Server{}.