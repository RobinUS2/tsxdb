#!/bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
for d in */ ; do
	if [ "staticcheck/" == $d ]; then
	  continue
	fi
	echo "$d"
	cd $d
	go fmt ./...
	go get -u ./...
	go mod tidy -v
	go vet ./...
  go test -cover -race -v ./...
	staticcheck -tests=false ./... # download from https://github.com/dominikh/go-tools/releases
	cd ..
done

echo "ALL PASSED"
