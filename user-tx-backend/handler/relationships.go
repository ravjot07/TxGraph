package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "user-tx-backend/models"
)

// GetUserRelationships handles GET /api/relationships/user/{id}
func (h *Handler) GetUserRelationships(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok {
        http.Error(w, "missing user id", http.StatusBadRequest)
        return
    }
    uid, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid user id", http.StatusBadRequest)
        return
    }

    user, conns, err := h.DB.GetUserRelationships(uid)
    if err != nil {
        http.Error(w, "fetch user relationships failed", http.StatusInternalServerError)
        return
    }

    resp := models.UserRelationships{
        User:        user,
        Connections: conns,
    }
    json.NewEncoder(w).Encode(resp)
}

// GetTransactionRelationships handles GET /api/relationships/transaction/{id}
func (h *Handler) GetTransactionRelationships(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr, ok := vars["id"]
    if !ok {
        http.Error(w, "missing transaction id", http.StatusBadRequest)
        return
    }
    txID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid transaction id", http.StatusBadRequest)
        return
    }

    txNode, conns, err := h.DB.GetTransactionRelationships(txID)
    if err != nil {
        http.Error(w, "fetch transaction relationships failed", http.StatusInternalServerError)
        return
    }

    resp := models.TransactionRelationships{
        Transaction: txNode,
        Connections: conns,
    }
    json.NewEncoder(w).Encode(resp)
}
