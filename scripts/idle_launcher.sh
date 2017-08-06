#!/bin/bash

while true; do
    for i in $(seq 1 200); do
        echo 'killing softly' $i
        docker kill -s USR1 idle-$i
        echo 'waiting for it to sleep' $i
        timeout 250 docker wait idle-$i
        echo 'killing it harder' $i
        docker kill idle-$i
        echo 'removing' $i
        docker rm idle-$i
        echo 'launching' $i
        make IDLENAME=idle-$i idle-run
    done
done
