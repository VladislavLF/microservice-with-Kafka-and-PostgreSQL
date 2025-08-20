package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"order-svc/internal/service"
)

type Handler struct {
	svc *service.Service
	log *zap.Logger
}

func NewHandler(s *service.Service, l *zap.Logger) *Handler {
	return &Handler{svc: s, log: l}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimPrefix(r.URL.Path, "/order/")
	if uid == "" {
		http.Error(w, "order id required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	o, err := h.svc.Get(ctx, uid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(o)
}
