#!/bin/bash
set -e

# build binary
GOOS=linux go build .

# build image
docker build --tag tsxdb-server .