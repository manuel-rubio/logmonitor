#!/bin/bash

for i in $(ls *.dot); do
    filename="${i%.*}"
    dot -Tpng $filename.dot > $filename.png
done

