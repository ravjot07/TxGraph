package handler

import (
    "encoding/json"
    "net/http"

    "user-tx-backend/models"
)

// CreateTransaction handles POST /api/transactions
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
    var req models.TransactionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    id, err := h.DB.CreateTransaction(
        req.FromUserID,
        req.ToUserID,
        req.Amount,
        req.Currency,
        req.Timestamp,
        req.Description,
        req.DeviceID,
    )
    if err != nil {
        http.Error(w, "create transaction failed", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// GetAllTransactions handles GET /api/transactions
func (h *Handler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
    txs, err := h.DB.GetAllTransactions()
    if err != nil {
        http.Error(w, "fetch transactions failed", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(txs)
}
