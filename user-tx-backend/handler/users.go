package handler

import (
    "encoding/json"
    "net/http"

    "user-tx-backend/graph"
    "user-tx-backend/models"
)

type Handler struct {
    DB *graph.Driver
}

func NewHandler(db *graph.Driver) *Handler {
    return &Handler{DB: db}
}

// CreateUser handles POST /api/users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req models.UserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    id, err := h.DB.CreateUser(req.Name, req.Email, req.Phone)
    if err != nil {
        http.Error(w, "create user failed", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// GetAllUsers handles GET /api/users
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.DB.GetAllUsers()
    if err != nil {
        http.Error(w, "fetch users failed", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(users)
}
