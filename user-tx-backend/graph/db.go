package graph

import (
    "context"
    "errors"
    "time"

    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
    "user-tx-backend/models"
)

type Driver struct {
    drv neo4j.DriverWithContext
}

// NewDriver creates a new Neo4j driver and verifies connectivity.
func NewDriver(uri, username, password string) (*Driver, error) {
    drv, err := neo4j.NewDriverWithContext(uri,
        neo4j.BasicAuth(username, password, ""),
    )
    if err != nil {
        return nil, err
    }
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := drv.VerifyConnectivity(ctx); err != nil {
        return nil, err
    }
    return &Driver{drv}, nil
}

func (d *Driver) Close() {
    _ = d.drv.Close(context.Background())
}

// CreateUser inserts a User node (no IP) and returns its internal Neo4j ID.
func (d *Driver) CreateUser(name, email, phone string) (int64, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    raw, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rec, err := tx.Run(ctx,
            `CREATE (u:User { name: $name, email: $email, phone: $phone })
             RETURN id(u)`,
            map[string]any{"name": name, "email": email, "phone": phone},
        )
        if err != nil {
            return nil, err
        }
        if rec.Next(ctx) {
            return rec.Record().Values[0].(int64), nil
        }
        return nil, errors.New("no record returned")
    })
    if err != nil {
        return 0, err
    }
    return raw.(int64), nil
}

// GetAllUsers retrieves all users (no IP in results).
func (d *Driver) GetAllUsers() ([]models.User, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    raw, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx,
            `MATCH (u:User)
             RETURN id(u) AS id, u.name AS name, u.email AS email, u.phone AS phone`,
            nil,
        )
        if err != nil {
            return nil, err
        }
        var users []models.User
        for result.Next(ctx) {
            rec := result.Record()
            users = append(users, models.User{
                ID:    rec.Values[0].(int64),
                Name:  rec.Values[1].(string),
                Email: rec.Values[2].(string),
                Phone: rec.Values[3].(string),
            })
        }
        return users, result.Err()
    })
    if err != nil {
        return nil, err
    }
    return raw.([]models.User), nil
}

// CreateTransaction inserts a Transaction node and links sender→transaction→receiver.
func (d *Driver) CreateTransaction(
    fromID, toID int64,
    amount float64,
    currency, timestamp, description, deviceID string,
) (int64, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    raw, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rec, err := tx.Run(ctx,
            `MATCH (u1:User),(u2:User)
             WHERE id(u1)=$fromId AND id(u2)=$toId
             CREATE (t:Transaction {
               amount:      $amt,
               currency:    $currency,
               timestamp:   datetime($ts),
               description: $desc,
               `+"`deviceId`"+`:    $dev
             })
             CREATE (u1)-[:SENT]->(t)
             CREATE (t)-[:RECEIVED_BY]->(u2)
             RETURN id(t)`,
            map[string]any{
                "fromId":   fromID,
                "toId":     toID,
                "amt":      amount,
                "currency": currency,
                "ts":       timestamp,
                "desc":     description,
                "dev":      deviceID,
            },
        )
        if err != nil {
            return nil, err
        }
        if rec.Next(ctx) {
            return rec.Record().Values[0].(int64), nil
        }
        return nil, errors.New("no record returned")
    })
    if err != nil {
        return 0, err
    }
    return raw.(int64), nil
}

// GetAllTransactions retrieves every transaction, including from/to IDs and deviceId.
func (d *Driver) GetAllTransactions() ([]models.Transaction, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    raw, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx,
            `MATCH (u1:User)-[:SENT]->(t:Transaction)-[:RECEIVED_BY]->(u2:User)
             RETURN id(t)           AS id,
                    id(u1)         AS fromId,
                    id(u2)         AS toId,
                    t.amount       AS amt,
                    t.currency     AS currency,
                    toString(t.timestamp) AS ts,
                    t.description  AS desc,
                    t.deviceId     AS deviceId`,
            nil,
        )
        if err != nil {
            return nil, err
        }
        var txs []models.Transaction
        for result.Next(ctx) {
            r := result.Record()
            txs = append(txs, models.Transaction{
                ID:          r.Values[0].(int64),
                FromUserID:  r.Values[1].(int64),
                ToUserID:    r.Values[2].(int64),
                Amount:      r.Values[3].(float64),
                Currency:    r.Values[4].(string),
                Timestamp:   r.Values[5].(string),
                Description: r.Values[6].(string),
                DeviceID:    r.Values[7].(string),
            })
        }
        return txs, result.Err()
    })
    if err != nil {
        return nil, err
    }
    return raw.([]models.Transaction), nil
}

// GetUserRelationships fetches a user plus both sent and received transactions.
func (d *Driver) GetUserRelationships(
    userID int64,
) (models.User, models.UserConnections, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    // 1) Fetch user
    var user models.User
    if _, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rec, err := tx.Run(ctx,
            `MATCH (u:User) WHERE id(u) = $uid
             RETURN id(u), u.name, u.email, u.phone`,
            map[string]any{"uid": userID},
        )
        if err != nil {
            return nil, err
        }
        if rec.Next(ctx) {
            r := rec.Record()
            user = models.User{
                ID:    r.Values[0].(int64),
                Name:  r.Values[1].(string),
                Email: r.Values[2].(string),
                Phone: r.Values[3].(string),
            }
            return nil, nil
        }
        return nil, errors.New("user not found")
    }); err != nil {
        return user, models.UserConnections{}, err
    }

    conns := models.UserConnections{}

    // 2) Shared‐attribute links (email & phone)
    if _, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx,
            `MATCH (u:User)-[r:SHARED_EMAIL|SHARED_PHONE]-(o:User)
             WHERE id(u) = $uid
             RETURN type(r), id(o), o.name, o.email, o.phone`,
            map[string]any{"uid": userID},
        )
        if err != nil {
            return nil, err
        }
        for result.Next(ctx) {
            r := result.Record()
            conns.Users = append(conns.Users, models.RelConnection[models.User]{
                Node: models.User{
                    ID:    r.Values[1].(int64),
                    Name:  r.Values[2].(string),
                    Email: r.Values[3].(string),
                    Phone: r.Values[4].(string),
                },
                Relationship: r.Values[0].(string),
            })
        }
        return nil, result.Err()
    }); err != nil {
        return user, conns, err
    }

    // 3) SENT transactions
    if _, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx,
            `MATCH (u:User)-[r:SENT]->(t:Transaction)-[:RECEIVED_BY]->(v:User)
             WHERE id(u) = $uid
             RETURN type(r), id(t), id(u), id(v),
                    t.amount, t.currency, toString(t.timestamp), t.description, t.deviceId`,
            map[string]any{"uid": userID},
        )
        if err != nil {
            return nil, err
        }
        for result.Next(ctx) {
            r := result.Record()
            conns.Transactions = append(conns.Transactions, models.RelConnection[models.Transaction]{
                Node: models.Transaction{
                    ID:          r.Values[1].(int64),
                    FromUserID:  r.Values[2].(int64),
                    ToUserID:    r.Values[3].(int64),
                    Amount:      r.Values[4].(float64),
                    Currency:    r.Values[5].(string),
                    Timestamp:   r.Values[6].(string),
                    Description: r.Values[7].(string),
                    DeviceID:    r.Values[8].(string),
                },
                Relationship: r.Values[0].(string),
            })
        }
        return nil, result.Err()
    }); err != nil {
        return user, conns, err
    }

    // 4) RECEIVED transactions
    if _, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx,
            `MATCH (x:User)-[:SENT]->(t:Transaction)-[r:RECEIVED_BY]->(u:User)
             WHERE id(u) = $uid
             RETURN type(r), id(t), id(x), id(u),
                    t.amount, t.currency, toString(t.timestamp), t.description, t.deviceId`,
            map[string]any{"uid": userID},
        )
        if err != nil {
            return nil, err
        }
        for result.Next(ctx) {
            r := result.Record()
            conns.Transactions = append(conns.Transactions, models.RelConnection[models.Transaction]{
                Node: models.Transaction{
                    ID:          r.Values[1].(int64),
                    FromUserID:  r.Values[2].(int64),
                    ToUserID:    r.Values[3].(int64),
                    Amount:      r.Values[4].(float64),
                    Currency:    r.Values[5].(string),
                    Timestamp:   r.Values[6].(string),
                    Description: r.Values[7].(string),
                    DeviceID:    r.Values[8].(string),
                },
                Relationship: r.Values[0].(string),
            })
        }
        return nil, result.Err()
    }); err != nil {
        return user, conns, err
    }

    return user, conns, nil
}

// GetTransactionRelationships fetches a transaction plus its sender and receiver.
func (d *Driver) GetTransactionRelationships(
    txID int64,
) (models.Transaction, models.TxConnections, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    // 1) Fetch just the transaction node itself
    var txNode models.Transaction
    if _, err := session.ExecuteRead(ctx, func(txn neo4j.ManagedTransaction) (any, error) {
        rec, err := txn.Run(ctx,
            `MATCH (t:Transaction)
             WHERE id(t) = $txid
             RETURN id(t), t.amount, t.currency,
                    toString(t.timestamp), t.description, t.deviceId`,
            map[string]any{"txid": txID},
        )
        if err != nil {
            return nil, err
        }
        if rec.Next(ctx) {
            r := rec.Record()
            txNode = models.Transaction{
                ID:          r.Values[0].(int64),
                Amount:      r.Values[1].(float64),
                Currency:    r.Values[2].(string),
                Timestamp:   r.Values[3].(string),
                Description: r.Values[4].(string),
                DeviceID:    r.Values[5].(string),
            }
            return nil, nil
        }
        return nil, errors.New("transaction not found")
    }); err != nil {
        return txNode, models.TxConnections{}, err
    }

    conns := models.TxConnections{}

    // 2) Sender connection
    if _, err := session.ExecuteRead(ctx, func(txn neo4j.ManagedTransaction) (any, error) {
        rec, err := txn.Run(ctx,
            `MATCH (u:User)-[r:SENT]->(t:Transaction)
             WHERE id(t) = $txid
             RETURN type(r), id(u), u.name, u.email, u.phone`,
            map[string]any{"txid": txID},
        )
        if err != nil {
            return nil, err
        }
        if rec.Next(ctx) {
            r := rec.Record()
            conns.Users = append(conns.Users, models.RelConnection[models.User]{
                Node: models.User{
                    ID:    r.Values[1].(int64),
                    Name:  r.Values[2].(string),
                    Email: r.Values[3].(string),
                    Phone: r.Values[4].(string),
                },
                Relationship: r.Values[0].(string),
            })
        }
        return nil, rec.Err()
    }); err != nil {
        return txNode, conns, err
    }

    // 3) Receiver connection
    if _, err := session.ExecuteRead(ctx, func(txnn neo4j.ManagedTransaction) (any, error) {
        rec, err := txnn.Run(ctx,
            `MATCH (t:Transaction)-[r:RECEIVED_BY]->(u:User)
             WHERE id(t) = $txid
             RETURN type(r), id(u), u.name, u.email, u.phone`,
            map[string]any{"txid": txID},
        )
        if err != nil {
            return nil, err
        }
        if rec.Next(ctx) {
            r := rec.Record()
            conns.Users = append(conns.Users, models.RelConnection[models.User]{
                Node: models.User{
                    ID:    r.Values[1].(int64),
                    Name:  r.Values[2].(string),
                    Email: r.Values[3].(string),
                    Phone: r.Values[4].(string),
                },
                Relationship: r.Values[0].(string),
            })
        }
        return nil, rec.Err()
    }); err != nil {
        return txNode, conns, err
    }

    return txNode, conns, nil
}

// SeedData populates sample users, shared‐attribute links, and transactions.
func SeedData(d *Driver) error {
    ctx := context.Background()

    // 1) Sample users
    sampleUsers := []struct {
        name, email, phone string
    }{
        {"Alice", "alice@example.com", "1111111111"},
        {"Bob",   "bob@example.com",   "2222222222"},
        {"Carol", "alice@example.com", "3333333333"}, // shares email with Alice
        {"Dave",  "dave@example.com",  "1111111111"}, // shares phone with Alice
        {"Eve",   "eve@example.com",   "2222222222"}, // shares phone with Bob
    }
    userIDs := make([]int64, len(sampleUsers))
    for i, u := range sampleUsers {
        id, err := d.CreateUser(u.name, u.email, u.phone)
        if err != nil {
            return err
        }
        userIDs[i] = id
    }

    // 2) Shared‐attribute relationships (email & phone)
    sharedRels := []struct {
        relType           string
        idxA, idxB        int
    }{
        {"SHARED_EMAIL", 0, 2}, 
        {"SHARED_PHONE", 0, 3}, 
        {"SHARED_PHONE", 1, 4}, 
    }
    for _, s := range sharedRels {
        session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
        _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
            _, err := tx.Run(ctx,
                `MATCH (a:User),(b:User)
                 WHERE id(a)=$a AND id(b)=$b
                 MERGE (a)-[:`+s.relType+`]-(b)`,
                map[string]any{
                    "a": userIDs[s.idxA],
                    "b": userIDs[s.idxB],
                },
            )
            return nil, err
        })
        session.Close(ctx)
        if err != nil {
            return err
        }
    }

    // 3) Sample transactions (with deviceId) covering various links
    txDefs := []struct {
        from, to     int
        amount       float64
        currency     string
        description  string
        deviceId     string
    }{
        {0, 1, 100.0, "USD", "Payment A→B", "dev-001"}, 
        {1, 2, 150.0, "USD", "Payment B→C", "dev-001"},
        {2, 0, 200.0, "USD", "Payment C→A", "dev-002"},
        {3, 4, 250.0, "USD", "Payment D→E", "dev-003"},
        {4, 3, 300.0, "USD", "Payment E→D", "dev-003"},
        {0, 3, 350.0, "USD", "Payment A→D", "dev-004"},
        {2, 4, 400.0, "USD", "Payment C→E", "dev-002"},
    }
    for i, t := range txDefs {
        ts := time.Now().Add(time.Duration(-i) * time.Hour).Format(time.RFC3339)
        if _, err := d.CreateTransaction(
            userIDs[t.from],
            userIDs[t.to],
            t.amount,
            t.currency,
            ts,
            t.description,
            t.deviceId,
        ); err != nil {
            return err
        }
    }

    return nil
}

// ShortestPathUsers finds the shortest path between two users (nodes of any type).
func (d *Driver) ShortestPathUsers(
    fromID, toID int64,
) ([]models.PathNode, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    raw, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx, 
            `MATCH (a:User),(b:User),
                   p = shortestPath((a)-[*]-(b))
             WHERE id(a)=$from AND id(b)=$to
             UNWIND nodes(p) AS n
             RETURN labels(n)[0] AS lbl,
                    id(n)          AS id,
                    CASE WHEN n:User THEN n.name ELSE '' END AS name`,
            map[string]any{"from": fromID, "to": toID},
        )        
        if err != nil {
            return nil, err
        }
        var path []models.PathNode
        for result.Next(ctx) {
            r := result.Record()
            path = append(path, models.PathNode{
                Type: r.Values[0].(string),
                ID:   r.Values[1].(int64),
                Name: r.Values[2].(string), // empty for Transaction
            })
        }
        if err := result.Err(); err != nil {
            return nil, err
        }
        if len(path) == 0 {
            return nil, errors.New("no path found")
        }
        return path, nil
    })
    if err != nil {
        return nil, err
    }
    return raw.([]models.PathNode), nil
}


// ExportGraph pulls every node and relationship for export.
func (d *Driver) ExportGraph() (models.GraphExportResponse, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    export := models.GraphExportResponse{}

    // 1) Nodes
    rawNodes, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rs, err := tx.Run(ctx,
            `MATCH (n)
             RETURN id(n) AS id,
                    labels(n)[0] AS type,
                    properties(n)    AS props`,
            nil,
        )
        if err != nil {
            return nil, err
        }
        var nodes []models.GraphNode
        for rs.Next(ctx) {
            rec := rs.Record()
            nodes = append(nodes, models.GraphNode{
                ID:         rec.Values[0].(int64),
                Type:       rec.Values[1].(string),
                Properties: rec.Values[2].(map[string]any),
            })
        }
        return nodes, rs.Err()
    })
    if err != nil {
        return export, err
    }
    export.Nodes = rawNodes.([]models.GraphNode)

    // 2) Relationships
    rawRels, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rs, err := tx.Run(ctx,
            `MATCH (a)-[r]->(b)
             RETURN id(a)             AS sourceId,
                    labels(a)[0]       AS sourceType,
                    type(r)            AS relationship,
                    id(b)             AS targetId,
                    labels(b)[0]      AS targetType`,
            nil,
        )
        if err != nil {
            return nil, err
        }
        var rels []models.GraphRelationship
        for rs.Next(ctx) {
            rec := rs.Record()
            rels = append(rels, models.GraphRelationship{
                SourceID:     rec.Values[0].(int64),
                SourceType:   rec.Values[1].(string),
                Relationship: rec.Values[2].(string),
                TargetID:     rec.Values[3].(int64),
                TargetType:   rec.Values[4].(string),
            })
        }
        return rels, rs.Err()
    })
    if err != nil {
        return export, err
    }
    export.Relationships = rawRels.([]models.GraphRelationship)

    // 3) Add TT edges for shared deviceId
    rawTT, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rs, err := tx.Run(ctx,
            `MATCH (t1:Transaction),(t2:Transaction)
             WHERE id(t1) < id(t2)
               AND (
                 (t1.ip IS NOT NULL AND t1.ip = t2.ip)
                  OR
                 (t1.deviceId IS NOT NULL AND t1.deviceId = t2.deviceId)
               )
             RETURN id(t1)                                   AS sourceId,
                    labels(t1)[0]                             AS sourceType,
                    CASE
                      WHEN t1.ip = t2.ip THEN 'SHARED_IP'
                      ELSE 'SHARED_DEVICE'
                    END                                         AS relationship,
                    id(t2)                                   AS targetId,
                    labels(t2)[0]                             AS targetType`,
            nil,
        )
        if err != nil {
            return nil, err
        }
        var ttRels []models.GraphRelationship
        for rs.Next(ctx) {
            rec := rs.Record()
            ttRels = append(ttRels, models.GraphRelationship{
                SourceID:     rec.Values[0].(int64),
                SourceType:   rec.Values[1].(string),
                Relationship: rec.Values[2].(string),
                TargetID:     rec.Values[3].(int64),
                TargetType:   rec.Values[4].(string),
            })
        }
        return ttRels, rs.Err()
    })
    if err != nil {
        return export, err
    }
    export.Relationships = append(export.Relationships, rawTT.([]models.GraphRelationship)...)

    return export, nil
}