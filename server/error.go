package server

import (
	"log/slog"
	"net/http"
)

func (s *Server) writeInternalError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "plain/text")
	// can't change the status after this call
	w.WriteHeader(http.StatusInternalServerError)

	s.logger.Error("internal server error", slog.String("error", err.Error()))

	_, _ = w.Write([]byte("internal server error: " + err.Error()))
}
