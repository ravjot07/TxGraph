#!/bin/sh
set -e

echo "⏳ Waiting for Neo4j to be ready..."
until nc -z -v -w30 neo4j 7687
do
  echo "Waiting for Neo4j at neo4j:7687..."
  sleep 3
done

echo "✅ Neo4j is up - launching backend..."
exec ./user-tx-backend
