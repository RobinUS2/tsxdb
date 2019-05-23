#!/bin/bash
set -e
LOGS_OPTS=${LOGS_OPTS:-"-f"}
docker logs ${LOGS_OPTS} `docker ps | grep 'tsxdb-server' | awk '{print $1}'`
