package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/llantera/hex/internal/domain/user"
)

// AuthHandler expone el endpoint de inicio de sesión basado en JWT.
type AuthHandler struct {
	service   user.Service
	jwtSecret []byte
	ttl       time.Duration
}

func NewAuthHandler(service user.Service, jwtSecret string, ttl time.Duration) *AuthHandler {
	if ttl <= 0 {
		// Valor por defecto: 6 días
		ttl = 24 * time.Hour * 6
	}
	return &AuthHandler{
		service:   service,
		jwtSecret: []byte(jwtSecret),
		ttl:       ttl,
	}
}

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authUserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Level        string `json:"level"`
	PriceLevelID *int   `json:"priceLevelId,omitempty"`
}

type loginResponse struct {
	Token     string           `json:"token"`
	ExpiresAt time.Time        `json:"expiresAt"`
	User      authUserResponse `json:"user"`
}

// HandleLogin procesa POST /api/auth/login.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var payload loginPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	email := strings.TrimSpace(strings.ToLower(payload.Email))
	password := strings.TrimSpace(payload.Password)
	if email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "correo y contraseña son obligatorios")
		return
	}

	u, err := h.service.Authenticate(r.Context(), email, password)
	if err != nil {
		switch err {
		case user.ErrInvalidCredentials:
			writeError(w, http.StatusUnauthorized, "Correo o contraseña incorrectos")
		default:
			var verr *user.ValidationError
			if errors.As(err, &verr) {
				writeError(w, http.StatusBadRequest, verr.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "error interno del servidor")
		}
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(h.ttl)

	claims := jwt.MapClaims{
		"sub":          u.ID,
		"email":        u.Email,
		"role":         string(u.Role),
		"level":        string(u.Level),
		"priceLevelId": u.PriceLevelID,
		"exp":          expiresAt.Unix(),
		"iat":          now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo generar el token de acceso")
		return
	}

	resp := loginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
		User: authUserResponse{
			ID:           u.ID,
			Email:        u.Email,
			Name:         u.Name,
			Role:         string(u.Role),
			Level:        string(u.Level),
			PriceLevelID: u.PriceLevelID,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}
