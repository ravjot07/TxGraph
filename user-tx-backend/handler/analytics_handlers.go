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
    toID, err2 := strconv.ParseInt(vars["to"], 10, 64)
    if err1 != nil || err2 != nil {
        http.Error(w, "invalid user IDs", http.StatusBadRequest)
        return
    }

    // Fetch the path segments (with from-node, to-node, relationship)
    segments, err := h.DB.ShortestPathSegments(fromID, toID)
    if err != nil {
        // If no path found, return 404, otherwise 500
        if err.Error() == "no path found" {
            http.Error(w, err.Error(), http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    // Encode the segments as JSON
    w.Header().Set("Content-Type", "application/json")
    resp := models.ShortestPathResponse{
        Segments: segments,
    }
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        http.Error(w, "failed to serialize response", http.StatusInternalServerError)
        return
    }
}

// GetTransactionClusters handles GET /api/analytics/transaction-clusters
func (h *Handler) GetTransactionClusters(w http.ResponseWriter, r *http.Request) {
    clusters, err := h.DB.ClusterTransactions()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(models.TransactionClustersResponse{
        Clusters: clusters,
    })
}