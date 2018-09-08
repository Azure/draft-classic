#!/usr/bin/env bash

REPO_ROOT=$PWD
EXAMPLES_FOLDER=$PWD/examples

for example in $(ls $EXAMPLES_FOLDER); do
    if [[ -d "$EXAMPLES_FOLDER/$example" ]]; then
        echo "Testing example $example"

        cd $EXAMPLES_FOLDER/$example
        rm -rf charts
        rm -rf Dockerfile
        
        draft create 
        draft up    
        draft delete 
    fi
done