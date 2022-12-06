package opendj_exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var commit string
var tag string

func GetVersion() string {
	return fmt.Sprintf("%s (%s)", tag, commit)
}

type Server struct {
	server *http.Server
	logger log.FieldLogger
}

func NewMetricsServer(bindAddr, metricsPath string) *Server {
	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.Handler())
	mux.HandleFunc("/version", showVersion)
	var err error
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write([]byte(`<html>
			<head><title>OpenDJ Exporter</title></head>
			<body>
			<h1>OpenDJ Exporter</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Fatalf("Error sending response body: %s", err)
		}
	})

	return &Server{
		server: &http.Server{Addr: bindAddr, Handler: mux},
		logger: log.WithField("component", "server"),
	}
}

func showVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, GetVersion())
}

func (s *Server) Start() error {
	s.logger.WithField("addr", s.server.Addr).Info("starting http listener")
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	s.server.Shutdown(ctx)
	cancel()
}
