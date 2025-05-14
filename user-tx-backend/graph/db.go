package graph

import (
    "context"
    "errors"
    "time"
    "fmt"

    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
    "user-tx-backend/models"
)

type Driver struct {
    drv neo4j.DriverWithContext
}

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

func (d *Driver) CreateUser(name, email, phone string) (int64, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    rawID, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
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
        return nil, fmt.Errorf("CreateUser: no record returned")
    })
    if err != nil {
        return 0, err
    }
    newID := rawID.(int64)

    if _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        _, err := tx.Run(ctx,
            `MATCH (u:User), (o:User)
             WHERE id(u) = $id AND o.email = u.email AND id(o) <> $id
             MERGE (u)-[:SHARED_EMAIL]-(o)`,
            map[string]any{"id": newID},
        )
        return nil, err
    }); err != nil {
        return newID, fmt.Errorf("CreateUser: failed to link SHARED_EMAIL: %w", err)
    }

    if _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        _, err := tx.Run(ctx,
            `MATCH (u:User), (o:User)
             WHERE id(u) = $id AND o.phone = u.phone AND id(o) <> $id
             MERGE (u)-[:SHARED_PHONE]-(o)`,
            map[string]any{"id": newID},
        )
        return nil, err
    }); err != nil {
        return newID, fmt.Errorf("CreateUser: failed to link SHARED_PHONE: %w", err)
    }

    return newID, nil
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
    currency, timestamp, description, deviceId string,
) (int64, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    rawID, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rec, err := tx.Run(ctx,
            `MATCH (u1:User),(u2:User)
             WHERE id(u1) = $fromId AND id(u2) = $toId
             CREATE (t:Transaction {
               amount:      $amt,
               currency:    $currency,
               timestamp:   datetime($ts),
               description: $desc,
               deviceId:    $deviceId
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
                "deviceId": deviceId,
            },
        )
        if err != nil {
            return nil, err
        }
        if rec.Next(ctx) {
            return rec.Record().Values[0].(int64), nil
        }
        return nil, fmt.Errorf("CreateTransaction: no record returned")
    })
    if err != nil {
        return 0, err
    }
    newID := rawID.(int64)

    if _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        _, err := tx.Run(ctx,
            `MATCH (t:Transaction),(o:Transaction)
             WHERE id(t) = $id
               AND o.deviceId = t.deviceId
               AND id(o) <> $id
             MERGE (t)-[:SHARED_DEVICE]-(o)`,
            map[string]any{"id": newID},
        )
        return nil, err
    }); err != nil {
        return newID, fmt.Errorf("CreateTransaction: failed to link SHARED_DEVICE: %w", err)
    }

    return newID, nil
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

func (d *Driver) ShortestPathSegments(
    fromID, toID int64,
) ([]models.PathSegment, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    raw, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx,
            `MATCH (a:User),(b:User), p = shortestPath((a)-[*]-(b))
             WHERE id(a) = $from AND id(b) = $to
             UNWIND relationships(p) AS r
             WITH r, startNode(r) AS fn, endNode(r) AS tn
             RETURN
               labels(fn)[0]                                        AS fromLabel,
               id(fn)                                               AS fromId,
               CASE WHEN fn:User THEN fn.name ELSE '' END           AS fromName,
               CASE WHEN fn:Transaction THEN fn.deviceId ELSE '' END AS fromDeviceId,

               labels(tn)[0]                                        AS toLabel,
               id(tn)                                               AS toId,
               CASE WHEN tn:User THEN tn.name ELSE '' END           AS toName,
               CASE WHEN tn:Transaction THEN tn.deviceId ELSE '' END AS toDeviceId,

               type(r)                                              AS relationship`,
            map[string]any{"from": fromID, "to": toID},
        )
        if err != nil {
            return nil, err
        }

        var segments []models.PathSegment
        for result.Next(ctx) {
            rec := result.Record()

            from := models.PathNode{
                Type:     rec.Values[0].(string),
                ID:       rec.Values[1].(int64),
                Name:     rec.Values[2].(string),
                DeviceID: rec.Values[3].(string),
            }
            to := models.PathNode{
                Type:     rec.Values[4].(string),
                ID:       rec.Values[5].(int64),
                Name:     rec.Values[6].(string),
                DeviceID: rec.Values[7].(string),
            }
            rel := rec.Values[8].(string)

            segments = append(segments, models.PathSegment{
                From:         from,
                To:           to,
                Relationship: rel,
            })
        }
        if err := result.Err(); err != nil {
            return nil, err
        }
        if len(segments) == 0 {
            return nil, errors.New("no path found")
        }
        return segments, nil
    })
    if err != nil {
        return nil, err
    }
    return raw.([]models.PathSegment), nil
}

func (d *Driver) ClusterTransactions() ([]models.TransactionCluster, error) {
    ctx := context.Background()
    session := d.drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    // 1) Load all transaction IDs
    rawTx, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rs, err := tx.Run(ctx, `MATCH (t:Transaction) RETURN id(t)`, nil)
        if err != nil {
            return nil, err
        }
        var ids []int64
        for rs.Next(ctx) {
            ids = append(ids, rs.Record().Values[0].(int64))
        }
        return ids, rs.Err()
    })
    if err != nil {
        return nil, err
    }
    txIDs := rawTx.([]int64)

    // 2) Load every transaction–transaction edge via a shared user
    rawPairs, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        rs, err := tx.Run(ctx, `
            MATCH (t1:Transaction)-[:SENT|RECEIVED_BY]-(u:User)-[:SENT|RECEIVED_BY]-(t2:Transaction)
            WHERE id(t1)<id(t2)
            RETURN id(t1), id(t2)
        `, nil)
        if err != nil {
            return nil, err
        }
        var pairs [][2]int64
        for rs.Next(ctx) {
            rec := rs.Record()
            pairs = append(pairs, [2]int64{
                rec.Values[0].(int64),
                rec.Values[1].(int64),
            })
        }
        return pairs, rs.Err()
    })
    if err != nil {
        return nil, err
    }
    transactionPairs := rawPairs.([][2]int64)

    // 3) Build adjacency list
    adj := make(map[int64][]int64, len(txIDs))
    for _, id := range txIDs {
        adj[id] = []int64{}
    }
    for _, p := range transactionPairs {
        t1, t2 := p[0], p[1]
        adj[t1] = append(adj[t1], t2)
        adj[t2] = append(adj[t2], t1)
    }

    // 4) BFS to find connected components
    visited := make(map[int64]bool, len(txIDs))
    var clusters []models.TransactionCluster

    for _, start := range txIDs {
        if visited[start] {
            continue
        }
        // collect this component
        queue := []int64{start}
        visited[start] = true
        comp := []int64{start}

        for len(queue) > 0 {
            curr := queue[0]
            queue = queue[1:]
            for _, nei := range adj[curr] {
                if !visited[nei] {
                    visited[nei] = true
                    queue = append(queue, nei)
                    comp = append(comp, nei)
                }
            }
        }

        // determine cluster ID = smallest transaction ID in comp
        minID := comp[0]
        for _, tid := range comp[1:] {
            if tid < minID {
                minID = tid
            }
        }

        // emit one entry per transaction in this component
        for _, tid := range comp {
            clusters = append(clusters, models.TransactionCluster{
                TransactionID: tid,
                ClusterID:     minID,
            })
        }
    }

    return clusters, nil
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