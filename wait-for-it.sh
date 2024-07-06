#!/bin/sh
# wait-for-it.sh

set -e

cmd="$1"

# Construct the connection string
conn_string="host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME password=$DB_PASSWORD sslmode=disable"


until PGPASSWORD=$DB_PASSWORD psql "$conn_string" -c "\q"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 5
done

>&2 echo "Postgres is up - executing command"
exec $cmd