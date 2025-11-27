package middleware

import (
	"net/http"
	"os"
	"strings"
)

func WithCORS(next http.Handler) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		// Lista de orígenes permitidos configurable por entorno.
		// Si la variable CORS_ALLOWED_ORIGINS está definida, se usa una lista
		// separada por comas. Si no, se usa la lista por defecto.
		allowedOrigins := map[string]bool{}
		if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
			for _, o := range strings.Split(envOrigins, ",") {
				trimmed := strings.TrimSpace(o)
				if trimmed != "" {
					allowedOrigins[trimmed] = true
				}
			}
		} else {
			allowedOrigins = map[string]bool{
				"https://ikingarisolorzano.com":        true,
				"https://www.ikingarisolorzano.com":    true,
				"https://ikingarisolorzano.com.mx":     true,
				"https://www.ikingarisolorzono.com.mx": true,

				// SSR Angular Universal
				"http://localhost:4000": true,
				"http://127.0.0.1:4000": true,
			}
		}

		// Validar si el origin está permitido
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // Importante para proxies
		}

		// Métodos y headers permitidos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Respuesta a OPTIONS sin llegar al handler final
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
