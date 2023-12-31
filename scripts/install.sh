#! /bin/bash

go build -o filer .

## Create symklink to /usr/local/bin
sudo ln -sf "$PWD"/filer /usr/local/bin/filer

filer completion zsh > _filer

sudo ln -sf "$PWD"/_filer "$(brew --prefix)"/share/zsh-autocomplete/_filer
