
# User & Transaction Dashboard

A web application for managing users and transactions with Neo4j graph relationships. Easily register users, record transactions, visualize connections, find shortest paths, cluster related transactions, and export the full graph.

## Features

-   **Add a User**: Create new users with name, email, and phone. Shared-email and shared-phone links are auto-generated.
    
-   **Add a Transaction**: Record payments between users, including amount, currency, timestamp, description, and device ID. Auto-link transactions sharing the same device.
    
-   **View Lists**: Browse all users and transactions in searchable, filterable tables.
    
-   **Graph View**: Interactive network visualization of users and their relationships (shared attributes, sent, received).
    
-   **Shortest Path Analytics**: Compute and display the shortest connection chain between any two users, with edge labels showing relationship types.
    
-   **Transaction Clusters**: Automatically cluster transactions that share common users, helping identify related activity.
    
-   **Export JSON/CSV**: Download the entire graph (nodes & relationships) in JSON or CSV formats.
    



## Architecture

-   **Backend**: Go + Neo4j driver, REST API endpoints under `/api/*`.
    
-   **Frontend**: React with React Router, Cytoscape.js for graph visualizations, Tailwind CSS for styling.
    
-   **Database**: Neo4j graph database.
    

## Quick Start

### Prerequisites

-   Docker & Docker Compose 
    
-   Go 1.20+
    
-   Node.js 16+
    
-   Neo4j Aura or local Neo4j 4.x+ instance
    

### Environment Variables

Create a `.env` file in the `user-tx-backend/` directory:

### For local setup

```
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASS=password123
SEED_DATA=true
PORT=8080
```
### For Docker setup

```
NEO4J_URI=bolt://neo4j:7687
NEO4J_USER=neo4j
NEO4J_PASS=your_password
SEED_DATA=true
PORT=8080
```
## Local Setup

### Backend

```bash
cd backend
go mod tidy

# build and run
go build -o backend .
./backend
```
### Database
#### Neo4j


```
docker run -d \
  --name neo4j \
  --network user-tx-net \
  -p7474:7474 \
  -p7687:7687 \
  -e NEO4J_AUTH=neo4j/password123 \
  -v neo4j-data:/data \
  neo4j:5.7.0
  ``` 

### Frontend

```bash
cd frontend
npm install
npm start
```

Open [http://localhost:3000](http://localhost:3000/) in your browser.

----------
## Running with standalone Docker

We’ll create a Docker network so that containers can talk by name:

```docker network create user-tx-net ```

### 1. Neo4j


```
docker run -d \
  --name neo4j \
  --network user-tx-net \
  -p7474:7474 \
  -p7687:7687 \
  -e NEO4J_AUTH=neo4j/password123 \
  -v neo4j-data:/data \
  neo4j:5.7.0
  ``` 

### 2. Go backend


```
   docker build -t user-tx-backend:latest ./backend # run docker run -d \
  --name user-tx-backend \
  --network user-tx-net \
  --env-file .env \
  -p 8080:8080 \
  user-tx-backend:latest
  ``` 

### 3. React frontend


```
docker build -t user-tx-frontend:latest ./user-tx-frontend # run docker run -d \
  --name user-tx-frontend \
  --network user-tx-net \
  -p 3000:80 \
  user-tx-frontend:latest
  ``` 

### 4. Verify

-   **Neo4j Browser:** [http://localhost:7474](http://localhost:7474) (login `neo4j`/`password123`)
    
-   **API:** [http://localhost:8080/api/users](http://localhost:8080/api/users)
    
-   **UI:** [http://localhost:3000](http://localhost:3000)



##  Running with Docker Compose

Compose will wire up the network and order of startup.


```
# stop & clean volumes 
docker compose down -v 
# build images & start all services in background 
docker compose up --build -d		
``` 

### 1. Check status

`docker compose ps` 

All three services (`neo4j`, `backend`, `frontend`) should be **Up**.

### 2. Tail logs (optional)


```
docker compose logs -f neo4j
docker compose logs -f backend
docker compose logs -f frontend
``` 

### 3. Verify in browser

-   **Neo4j:** [http://localhost:7474](http://localhost:7474)
    
-   **API:** [http://localhost:8080/api/users](http://localhost:8080/api/users)
    
-   **UI:** [http://localhost:3000](http://localhost:3000)
    


##  Stopping

When you’re done, bring everything down and remove the data:

```
docker compose down -v
docker network rm user-tx-net 
```



## API Endpoints

```markdown                               
|---------------|------------------------------------------------|---------------------------------------|
|     Method    |                    Endpoint                    |              Description              |   
|---------------|------------------------------------------------|---------------------------------------|
| POST          | /api/users                                     | Create a new user                     |   
| GET           | /api/users                                     | List all users                        |   
| POST          | /api/transactions                              | Create a new transaction              |   
| GET           | /api/transactions                              | List all transactions                 |   
| GET           | /api/relationships/user/{id}                   | Get user relationships (graph branch) |   
| GET           | /api/relationships/transaction/{id}            | Get transaction relationships         |   
| GET           | /api/analytics/shortest-path/users/{from}/{to} | Shortest path between two users       |   
| GET           | /api/analytics/transaction-clusters            | Cluster transactions by shared users  |   
| GET           | /api/export/json                               | Export entire graph as JSON           |   
| GET           | /api/export/csv                                | Export entire graph as CSV            |   
```