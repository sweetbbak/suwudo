#!/usr/bin/env bash

! command -v go && {
    echo "Go is not installed"
    exit
}

printf "\e[3;33;3m%s\e[0m\n" "Building suwudo"
go build
sudo chown root suwu
sudo chmod u+s suwu
