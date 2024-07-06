#!/bin/sh
# wait-for-it.sh

set -e

cmd="$1"

# Debugging information
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "Command: $cmd"
echo "DB User: $DB_USER"
echo "DB Name: $DB_NAME"
echo "DB PASSWORD: $DB_PASSWORD"

# Construct the connection string
conn_string="host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME password=$DB_PASSWORD sslmode=disable"


until PGPASSWORD=$DB_PASSWORD psql "$conn_string" -c "\q"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 5
done

>&2 echo "Postgres is up - executing command"
exec $cmd