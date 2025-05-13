package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "user-tx-backend/models"
)

// GetUserShortestPath handles GET /api/analytics/shortest-path/users/{from}/{to}
func (h *Handler) GetUserShortestPath(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fromID, err1 := strconv.ParseInt(vars["from"], 10, 64)
    toID,   err2 := strconv.ParseInt(vars["to"],   10, 64)
    if err1 != nil || err2 != nil {
        http.Error(w, "invalid user IDs", http.StatusBadRequest)
        return
    }

    path, err := h.DB.ShortestPathUsers(fromID, toID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    resp := models.ShortestPathResponse{Path: path}
    json.NewEncoder(w).Encode(resp)
}

