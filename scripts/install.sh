#! /bin/bash

go build -o fileb .

## Create symklink to /usr/local/bin
sudo ln -sf "$PWD"/fileb /usr/local/bin/fileb
