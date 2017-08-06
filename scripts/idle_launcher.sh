#!/bin/bash

for i in $(seq 1 10); do
    echo 'launching' $i
    make IDLENAME=idle-$i idle-run
done
