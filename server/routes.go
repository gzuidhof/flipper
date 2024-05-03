package server

import (
	"net/http"

	"github.com/gzuidhof/flipper/buildinfo"
	"github.com/gzuidhof/flipper/view/template"
)

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	s.mux.HandleFunc("GET /v1/version", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(buildinfo.Version()))
	})

	staticHandler := http.FileServerFS(s.staticFS)

	s.mux.Handle("GET /static/{path...}", http.StripPrefix("/static/", staticHandler))
	s.mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		err := template.RenderTemplateHTML(s.template.HomePage(), w, template.HomePageData{})
		if err != nil {
			s.writeInternalError(w, err)
			return
		}
	}))

	// 404 handler
	// s.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	w.Write([]byte("not found"))
	// })
}
