#!/bin/bash
set -e
for d in */ ; do
	echo "$d"
	cd $d
	go test -cover -race ./...
	cd ..
done
