#!/bin/bash

while true; do
    for i in $(seq 1 200); do
        echo 'killing' $i
        docker kill idle-$i
        echo 'removing' $i
        docker rm idle-$i
        echo 'building' $i
        make build
        echo 'launching' $i
        make IDLENAME=idle-$i idle-run
    done
done
