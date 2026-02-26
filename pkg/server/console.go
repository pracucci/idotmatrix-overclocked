package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

func (s *Server) registerConsoleRoutes(mux *http.ServeMux) {
	// Serve embedded static files
	subFS, _ := fs.Sub(assets, "assets")
	fileServer := http.FileServer(http.FS(subFS))
	mux.Handle("/", fileServer)
}
