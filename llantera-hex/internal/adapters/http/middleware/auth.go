package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Context keys para información de autenticación.
type contextKey string

const (
	userIDKey         contextKey = "authUserID"
	userEmailKey      contextKey = "authUserEmail"
	userRoleKey       contextKey = "authUserRole"
	userLevelKey      contextKey = "authUserLevel"
	userPriceLevelKey contextKey = "authUserPriceLevelId"
)

// Claims expone los datos principales del usuario extraídos del token.
type Claims struct {
	UserID       string
	Email        string
	Role         string
	Level        string
	PriceLevelID *int
}

// FromContext permite recuperar los claims desde un *http.Request.
func FromContext(ctx context.Context) *Claims {
	id, _ := ctx.Value(userIDKey).(string)
	if id == "" {
		return nil
	}
	email, _ := ctx.Value(userEmailKey).(string)
	role, _ := ctx.Value(userRoleKey).(string)
	level, _ := ctx.Value(userLevelKey).(string)
	var priceLevelID *int
	if v, ok := ctx.Value(userPriceLevelKey).(int); ok {
		priceLevelID = &v
	}
	return &Claims{
		UserID:       id,
		Email:        email,
		Role:         role,
		Level:        level,
		PriceLevelID: priceLevelID,
	}
}

// WithAuth valida el JWT en el header Authorization y agrega los claims al contexto.
// Si el token es inválido o falta, responde 401.
func WithAuth(secret string, next http.Handler) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Asegurar algoritmo esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return []byte(secret), nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimSpace(r.Header.Get("Authorization"))
		if raw == "" || !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("{\"error\":\"token de acceso requerido\"}"))
			return
		}

		tokenStr := strings.TrimSpace(raw[len("Bearer "):])
		if tokenStr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("{\"error\":\"token de acceso requerido\"}"))
			return
		}

		token, err := jwt.Parse(tokenStr, keyFunc)
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("{\"error\":\"token de acceso inválido\"}"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("{\"error\":\"token de acceso inválido\"}"))
			return
		}

		// Validación básica de expiración (además del chequeo interno de jwt.Parse).
		if expVal, ok := claims["exp"]; ok {
			switch v := expVal.(type) {
			case float64:
				if time.Now().Unix() > int64(v) {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte("{\"error\":\"token de acceso expirado\"}"))
					return
				}
			}
		}

		userID, _ := claims["sub"].(string)
		if strings.TrimSpace(userID) == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("{\"error\":\"token de acceso inválido\"}"))
			return
		}

		email, _ := claims["email"].(string)
		role, _ := claims["role"].(string)
		level, _ := claims["level"].(string)

		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		ctx = context.WithValue(ctx, userEmailKey, email)
		ctx = context.WithValue(ctx, userRoleKey, strings.ToLower(strings.TrimSpace(role)))
		ctx = context.WithValue(ctx, userLevelKey, level)

		if v, ok := claims["priceLevelId"]; ok {
			switch vv := v.(type) {
			case float64:
				ctx = context.WithValue(ctx, userPriceLevelKey, int(vv))
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID obtiene el ID del usuario autenticado como string (UUID)
func GetUserID(ctx context.Context) string {
	idStr, _ := ctx.Value(userIDKey).(string)
	return idStr
}

// GetUserRole obtiene el rol del usuario autenticado
func GetUserRole(ctx context.Context) string {
	role, _ := ctx.Value(userRoleKey).(string)
	return role
}

// WithRole restringe el acceso a uno o más roles específicos. Debe usarse
// después de WithAuth (es decir, sobre un handler ya protegido con WithAuth).
func WithRole(allowedRoles []string, next http.Handler) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}

	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		key := strings.ToLower(strings.TrimSpace(r))
		if key != "" {
			allowed[key] = struct{}{}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(userRoleKey).(string)
		role = strings.ToLower(strings.TrimSpace(role))
		if _, ok := allowed[role]; !ok {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("{\"error\":\"acceso no autorizado para este rol\"}"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
