package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler sirve archivos estáticos desde el directorio de trabajo
type FileHandler struct {
	baseDir string
}

// NewFileHandler crea un nuevo handler de archivos
func NewFileHandler(baseDir string) *FileHandler {
	return &FileHandler{baseDir: baseDir}
}

// ServeFile sirve un archivo desde el directorio base
func (h *FileHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer la ruta del archivo desde /api/files/
	// Ejemplo: /api/files/uploads/invoices/7/factura.pdf -> uploads/invoices/7/factura.pdf
	filePath := strings.TrimPrefix(r.URL.Path, "/api/files/")

	// Prevenir path traversal
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// La ruta ya incluye "uploads/", así que usamos directamente
	fullPath := cleanPath

	log.Printf("FileHandler: Solicitado archivo: %s -> %s", r.URL.Path, fullPath)

	// Verificar que el archivo existe
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Printf("FileHandler: Archivo no encontrado: %s", fullPath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Determinar content-type basado en extensión
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".pdf":
		w.Header().Set("Content-Type", "application/pdf")
	case ".xml":
		w.Header().Set("Content-Type", "application/xml")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	log.Printf("FileHandler: Sirviendo archivo: %s", fullPath)

	// Servir el archivo
	http.ServeFile(w, r, fullPath)
}
