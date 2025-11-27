package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/llantera/hex/internal/domain/tire"
)

// TireBrandHandler expone operaciones HTTP para el catálogo de marcas de llantas.
type TireBrandHandler struct {
	repo tire.BrandRepository
}

func NewTireBrandHandler(repo tire.BrandRepository) *TireBrandHandler {
	return &TireBrandHandler{repo: repo}
}

// brandAdminDTO representa la vista de administración de una marca con sus aliases.
type brandAdminDTO struct {
	ID            int      `json:"id"`
	Nombre        string   `json:"nombre"`
	Aliases       []string `json:"aliases"`
	CreadoEn      string   `json:"creadoEn"`
	ActualizadoEn string   `json:"actualizadoEn"`
}

// HandleCollection atiende las rutas /api/brands/.
func (h *TireBrandHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// HandleResource atiende rutas /api/brands/{id} para operaciones sobre una marca específica.
func (h *TireBrandHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	// Esperamos paths del tipo /api/brands/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		writeError(w, http.StatusNotFound, "recurso no encontrado")
		return
	}
	idStr := parts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "identificador de marca inválido")
		return
	}

	switch r.Method {
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		w.Header().Set("Allow", "PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

func (h *TireBrandHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brands, err := h.repo.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudieron obtener las marcas")
		return
	}

	result := make([]brandAdminDTO, 0, len(brands))
	for _, b := range brands {
		aliases, err := h.repo.ListAliases(ctx, b.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron obtener los alias de marcas")
			return
		}
		result = append(result, brandAdminDTO{
			ID:            b.ID,
			Nombre:        b.Nombre,
			Aliases:       aliases,
			CreadoEn:      b.CreadoEn.Format(time.RFC3339),
			ActualizadoEn: b.ActualizadoEn.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": result,
		"meta": map[string]int{
			"total": len(result),
		},
	})
}

type brandPayload struct {
	Nombre  string   `json:"nombre"`
	Aliases []string `json:"aliases"`
}

func (h *TireBrandHandler) decodePayload(r *http.Request) (brandPayload, error) {
	var payload brandPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		return brandPayload{}, err
	}
	return payload, nil
}

func (h *TireBrandHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	payload, err := h.decodePayload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	name := strings.TrimSpace(payload.Nombre)
	if name == "" {
		writeError(w, http.StatusBadRequest, "el nombre de la marca es obligatorio")
		return
	}

	marca := &tire.Brand{Nombre: name}
	if err := h.repo.Create(ctx, marca, payload.Aliases); err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo crear la marca")
		return
	}

	aliases, err := h.repo.ListAliases(ctx, marca.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudieron obtener los alias de la marca recién creada")
		return
	}

	resp := brandAdminDTO{
		ID:            marca.ID,
		Nombre:        marca.Nombre,
		Aliases:       aliases,
		CreadoEn:      marca.CreadoEn.Format(time.RFC3339),
		ActualizadoEn: marca.ActualizadoEn.Format(time.RFC3339),
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *TireBrandHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	ctx := r.Context()
	payload, err := h.decodePayload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	name := strings.TrimSpace(payload.Nombre)
	if name == "" {
		writeError(w, http.StatusBadRequest, "el nombre de la marca es obligatorio")
		return
	}

	marca, err := h.repo.GetByID(ctx, id)
	if err != nil {
		if err == tire.ErrBrandNotFound {
			writeError(w, http.StatusNotFound, "marca no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "no se pudo obtener la marca")
		return
	}

	marca.Nombre = name
	if err := h.repo.Update(ctx, marca, payload.Aliases); err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo actualizar la marca")
		return
	}

	aliases, err := h.repo.ListAliases(ctx, marca.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudieron obtener los alias de la marca actualizada")
		return
	}

	resp := brandAdminDTO{
		ID:            marca.ID,
		Nombre:        marca.Nombre,
		Aliases:       aliases,
		CreadoEn:      marca.CreadoEn.Format(time.RFC3339),
		ActualizadoEn: marca.ActualizadoEn.Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *TireBrandHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	ctx := r.Context()

	hasTires, err := h.repo.HasTires(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo validar si la marca tiene llantas asociadas")
		return
	}
	if hasTires {
		writeError(w, http.StatusBadRequest, "no se puede eliminar una marca que tiene llantas asociadas")
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo eliminar la marca")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
