#!/usr/bin/env bash

res=0

for i in `find . \( -path ./vendor -o -path ./vpp -o -path dir3 \) -prune -o -name "*.md"`
do
    if [ -d "$i" ]; then
        continue
    fi

    echo $i;
    if ! markdown-link-check $i; then
        res=1
    fi
    echo "";
done

exit $res