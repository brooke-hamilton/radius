#!/bin/sh
# ------------------------------------------------------------
# Copyright 2023 The Radius Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ------------------------------------------------------------
#
# Initializes the PostgreSQL databases used by the Radius control plane.
# Creates the "ucp" and "applications_rp" roles/databases and the "resources"
# table in each. Idempotent: safe to run repeatedly.
#
# Mirrors the database initialization performed by build/scripts/start-radius.sh.
set -eu

: "${POSTGRES_PASSWORD:=radius_pass}"
ADMIN_URL="postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/postgres?sslmode=disable"

echo "Waiting for PostgreSQL to accept connections..."
until psql "${ADMIN_URL}" -c "SELECT 1;" >/dev/null 2>&1; do
  sleep 2
done
echo "PostgreSQL is ready."

# Create roles and databases (idempotent via \gexec).
psql "${ADMIN_URL}" -v ON_ERROR_STOP=1 <<'SQL'
SELECT 'CREATE ROLE ucp LOGIN PASSWORD ''radius_pass'''
  WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'ucp')\gexec
SELECT 'CREATE ROLE applications_rp LOGIN PASSWORD ''radius_pass'''
  WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'applications_rp')\gexec
SELECT 'CREATE DATABASE ucp OWNER ucp'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ucp')\gexec
SELECT 'CREATE DATABASE applications_rp OWNER applications_rp'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'applications_rp')\gexec
SQL

# Create the resources table and grants in a database. Args: <db> <owner>
init_db_tables() {
  db="$1"
  owner="$2"
  db_url="postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/${db}?sslmode=disable"
  echo "Initializing tables in database '${db}'..."
  psql "${db_url}" -v ON_ERROR_STOP=1 <<SQL
CREATE TABLE IF NOT EXISTS resources (
  id TEXT PRIMARY KEY NOT NULL,
  original_id TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  root_scope TEXT NOT NULL,
  routing_scope TEXT NOT NULL,
  etag TEXT NOT NULL,
  created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP,
  resource_data jsonb NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_resource_query ON resources (resource_type, root_scope);
GRANT ALL PRIVILEGES ON TABLE resources TO ${owner};
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${owner};
SQL
}

init_db_tables ucp ucp
init_db_tables applications_rp applications_rp

echo "Database bootstrap complete."
