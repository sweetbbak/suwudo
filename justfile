alias b := default
alias build := default

default:
    printf "\e[3;33;3m%s\e[0m\n" "Building suwudo"
    go build
    sudo chown root suwu
    sudo chmod u+s suwu
    ./suwu

install:
    printf "\e[3;33;3m%s\e[0m\n" "Installing suwudo"
    sudo /usr/bin/cp ./suwu /usr/bin

test:
    go build
    sudo chown root suwu
    sudo chmod u+s suwu
    ./suwu cat /etc/shadow
    ./suwu su
