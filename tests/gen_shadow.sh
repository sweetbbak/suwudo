#!/usr/bin/env bash
# generate a dummy /etc/shadow file

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

fake_users=(
    "test123"
    "xyztest"
    "suwudo"
    "root"
)

fake_users2=(
    "go"
    "NIxLKJ97f"
    "suw udo"
    "rooot"
    "testt"
)

for x in "${fake_users[@]}"; do
    pass=$(openssl passwd -6 -salt "$x" $RANDOM)
    printf "$x:%s:19699:0:99999:7:::\n" "$pass" >> "$SCRIPT_DIR/shadow"
done

pass=$(openssl passwd -6 -salt xyz hunter2)
printf "test:%s:19699:0:99999:7:::\n" "$pass" >> "$SCRIPT_DIR/shadow"

for x in "${fake_users2[@]}"; do
    pass=$(openssl passwd -6 -salt "$x" $RANDOM)
    printf "$x:%s:19699:0:99999:7:::\n" "$pass" >> "$SCRIPT_DIR/shadow"
done
