package metric

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ryoeuyo/auth-microservice/internal/config"
	"net"
	"net/http"
	"strconv"
)

type Server struct {
	address string
	port    uint16
}

func NewServer(cfg config.MetricServer) *Server {
	return &Server{
		address: cfg.Address,
		port:    cfg.Port,
	}
}

func (s *Server) MustStart() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(net.JoinHostPort(s.address, strconv.Itoa(int(s.port))), mux); err != nil {
		panic(err)
	}
}