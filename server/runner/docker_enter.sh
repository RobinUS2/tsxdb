#!/bin/bash
set -e
docker exec -it `docker ps | grep 'tsxdb-server' | awk '{print $1}'` bash
