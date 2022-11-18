#!/bin/bash
set -e
docker run -d -p 1234:1234 -p 5555:5555 --restart unless-stopped --name tsxdb-server tsxdb-server
