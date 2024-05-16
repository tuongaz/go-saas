package frontend

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

//go:embed dist/*
var WebFiles embed.FS

func Handler(router chi.Router, files embed.FS, dir string) error {
	fileSystem, err := fs.Sub(files, dir)
	if err != nil {
		return fmt.Errorf("create file system: %w", err)
	}

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		file, err := fileSystem.Open(path)
		if err != nil {
			// File does not exist, serve index.html
			indexContent, err := fs.ReadFile(fileSystem, "index.html")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(indexContent))
		} else {
			defer file.Close()
			// Serve the requested file as it exists
			http.FileServer(http.FS(fileSystem)).ServeHTTP(w, r)
		}
	})

	return nil
}
