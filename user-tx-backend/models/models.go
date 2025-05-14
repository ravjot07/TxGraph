package models

// User represents a graph User node.
type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Phone string `json:"phone"`
}

// Transaction represents a graph Transaction node.
type Transaction struct {
    ID          int64   `json:"id"`
    FromUserID  int64   `json:"fromUserId"`
    ToUserID    int64   `json:"toUserId"`
    Amount      float64 `json:"amount"`
    Currency    string  `json:"currency"`
    Timestamp   string  `json:"timestamp"`
    Description string  `json:"description"`
    DeviceID    string  `json:"deviceId"`
}

// RelConnection wraps any node with its relationship type.
type RelConnection[T any] struct {
    Node         T      `json:"node"`
    Relationship string `json:"relationship"`
}

// UserConnections groups shared‐attribute and performed links.
type UserConnections struct {
    Users        []RelConnection[User]        `json:"users"`
    Transactions []RelConnection[Transaction] `json:"transactions"`
}

// TxConnections groups users who performed a transaction.
type TxConnections struct {
    Users []RelConnection[User] `json:"users"`
}

// UserRequest for POST /users
type UserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Phone string `json:"phone"`
}

// TransactionRequest for POST /api/transactions
type TransactionRequest struct {
    FromUserID  int64   `json:"fromUserId"`
    ToUserID    int64   `json:"toUserId"`
    Amount      float64 `json:"amount"`
    Currency    string  `json:"currency"`
    Timestamp   string  `json:"timestamp"`
    Description string  `json:"description"`
    DeviceID    string  `json:"deviceId"`
}

// Response wrappers for relationships

// UserRelationships used in GET /api/relationships/user/{id}
type UserRelationships struct {
    User        User            `json:"user"`
    Connections UserConnections `json:"connections"`
}

// TransactionRelationships used in GET /api/relationships/transaction/{id}
type TransactionRelationships struct {
    Transaction Transaction   `json:"transaction"`
    Connections TxConnections `json:"connections"`
}

type PathNode struct {
    ID       int64  `json:"id"`
    Type     string `json:"type"`           
    Name     string `json:"name,omitempty"` 
    DeviceID string `json:"deviceId,omitempty"`
}

// represent each hop in the response
type PathSegment struct {
    From         PathNode `json:"from"`
    To           PathNode `json:"to"`
    Relationship string   `json:"relationship"`
}

// response wrapper
type ShortestPathResponse struct {
    Segments []PathSegment `json:"segments"`
}

// GraphNode represents any node (User or Transaction) for export.
type GraphNode struct {
    ID         int64             `json:"id"`
    Type       string            `json:"type"`       
    Properties map[string]any    `json:"properties"` 
}

// GraphRelationship represents an edge in the graph.
type GraphRelationship struct {
    SourceID     int64  `json:"sourceId"`
    SourceType   string `json:"sourceType"`
    Relationship string `json:"relationship"`
    TargetID     int64  `json:"targetId"`
    TargetType   string `json:"targetType"`
}

// GraphExportResponse wraps the whole graph.
type GraphExportResponse struct {
    Nodes         []GraphNode         `json:"nodes"`
    Relationships []GraphRelationship `json:"relationships"`
}

// TransactionCluster represents a single transaction’s cluster assignment.
type TransactionCluster struct {
    TransactionID int64 `json:"transactionId"`
    ClusterID     int64 `json:"clusterId"`
}

// TransactionClustersResponse wraps all cluster assignments.
type TransactionClustersResponse struct {
    Clusters []TransactionCluster `json:"clusters"`
}
