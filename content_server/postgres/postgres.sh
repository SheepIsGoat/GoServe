#!/bin/bash

CONTAINER_NAME="my-sheep-go-postgres"
# Get the directory where this script is located
SCRIPT_DIR=$(realpath $(dirname "$0"))
echo "Mounting scripts from $SCRIPT_DIR"

# Append the relative path to your SQL script
SQL_SCRIPT_DIR="${SCRIPT_DIR}/sql"


EXISTS=$(docker ps -aq -f name="$CONTAINER_NAME")
if [ -z "$EXISTS" ]; then
    # Docker command
    docker run \
        --name "$CONTAINER_NAME" \
        -v "${SQL_SCRIPT_DIR}:/docker-entrypoint-initdb.d" \
        -e POSTGRES_PASSWORD=mysecretpassword \
        -d -p 5432:5432 postgres
else
    docker start "$CONTAINER_NAME"
fi