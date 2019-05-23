#!/bin/bash
set -e
docker run -d -p 1234 -p 5555 --name tsxdb-server tsxdb-server
