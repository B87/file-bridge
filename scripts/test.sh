#!/bin/bash 

CODACY_PUSH=$1

mkdir -p .coverage
if ! go test -v -coverprofile=.coverage/cover.out ./...; then
   echo "Test failed."
   exit 1
fi

if [ "$CODACY_PUSH" = "true" ]; then
    echo "Codacy push enabled"
    if [ -z "$CODACY_PROJECT_TOKEN" ]; then
        echo "Codacy push enabled but token not set"
        exit 0
    else
        bash <(curl -Ls https://coverage.codacy.com/get.sh) report --force-coverage-parser go -r .coverage/cover.out
    fi
else
    echo "Codacy push disabled"
    exit 0
fi