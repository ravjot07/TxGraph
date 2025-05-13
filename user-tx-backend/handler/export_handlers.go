package handler

import (
    "encoding/csv"
    "encoding/json"
    "net/http"
    "strconv"

)

// ExportGraphJSON handles GET /api/export/json
func (h *Handler) ExportGraphJSON(w http.ResponseWriter, r *http.Request) {
    data, err := h.DB.ExportGraph()
    if err != nil {
        http.Error(w, "export failed: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

// ExportGraphCSV handles GET /api/export/csv
func (h *Handler) ExportGraphCSV(w http.ResponseWriter, r *http.Request) {
    data, err := h.DB.ExportGraph()
    if err != nil {
        http.Error(w, "export failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename=\"graph.csv\"")
    writer := csv.NewWriter(w)

    // Nodes section
    writer.Write([]string{"# Nodes"})
    writer.Write([]string{"id", "type", "properties"})
    for _, n := range data.Nodes {
        props, _ := json.Marshal(n.Properties)
        writer.Write([]string{
            strconv.FormatInt(n.ID, 10),
            n.Type,
            string(props),
        })
    }
    writer.Write([]string{}) 

    // Relationships section
    writer.Write([]string{"# Relationships"})
    writer.Write([]string{"source_id", "source_type", "relationship", "target_id", "target_type"})
    for _, r := range data.Relationships {
        writer.Write([]string{
            strconv.FormatInt(r.SourceID, 10),
            r.SourceType,
            r.Relationship,
            strconv.FormatInt(r.TargetID, 10),
            r.TargetType,
        })
    }
    writer.Flush()
}
