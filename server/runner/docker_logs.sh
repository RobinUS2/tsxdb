#!/bin/bash
set -e
docker logs -f `docker ps | grep 'tsxdb-server' | awk '{print $1}'`
