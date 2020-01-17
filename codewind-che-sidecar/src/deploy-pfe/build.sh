#! /bin/sh

CGO_ENABLED=0 go build -o deploy-pfe -ldflags="-s -w"