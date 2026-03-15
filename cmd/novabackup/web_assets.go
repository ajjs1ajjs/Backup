package main

import (
	"embed"
	"io/fs"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Embedded assets are sourced from cmd/novabackup/web.
// Keep this folder in sync with the top-level web directory.
//
//go:embed web/*
var embeddedWebFS embed.FS

var embeddedWebSubFS fs.FS

func init() {
	sub, err := fs.Sub(embeddedWebFS, "web")
	if err == nil {
		embeddedWebSubFS = sub
	}
}

func serveWebFile(c *gin.Context, name string) {
	cleaned := path.Clean("/" + name)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" || cleaned == "." {
		cleaned = "index.html"
	}

	if webDir != "" {
		diskPath := filepath.Join(webDir, filepath.FromSlash(cleaned))
		if fileExists(diskPath) {
			c.File(diskPath)
			return
		}
	}

	if embeddedWebSubFS != nil {
		if data, err := fs.ReadFile(embeddedWebSubFS, cleaned); err == nil {
			c.Data(200, contentTypeFor(cleaned), data)
			return
		}
	}

	c.JSON(404, gin.H{"error": "Page not found"})
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func contentTypeFor(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "application/octet-stream"
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	default:
		return "application/octet-stream"
	}
}
