#!/bin/bash

git pull

soda migrate

go build -o Go-B-and-B ./cmd/web

sudo supervisorctl restart bnb

echo "Done"
