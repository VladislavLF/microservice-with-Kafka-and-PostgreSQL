package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"L0/internal/cache"
)

type Handler struct {
	cache   *cache.Cache
	webRoot string
}

func New(c *cache.Cache, webRoot string) *Handler {
	return &Handler{
		cache:   c,
		webRoot: webRoot,
	}
}

func (h *Handler) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	staticDir := filepath.Join(h.webRoot, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	mux.HandleFunc("/order/", h.handleOrderRequest)

	mux.HandleFunc("/", h.handleFrontend)

	return mux
}

func (h *Handler) handleFrontend(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	indexPath := filepath.Join(h.webRoot, "templates", "index.html")
	http.ServeFile(w, r, indexPath)
}

func (h *Handler) handleOrderRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uid := r.URL.Path[len("/order/"):]
	if uid == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	order, exists := h.cache.Get(uid)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
