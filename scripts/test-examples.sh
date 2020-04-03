#!/usr/bin/env bash

REPO_ROOT=$PWD
EXAMPLES_FOLDER=$PWD/examples
SKIP_K8S=${1:-false}

for example in $(ls $EXAMPLES_FOLDER); do
    if [[ -d "$EXAMPLES_FOLDER/$example" ]]; then
        echo "Testing example $example"

        cd $EXAMPLES_FOLDER/$example
        rm -rf charts
        rm -rf Dockerfile
        
        draft create
        if [ $SKIP_K8S = true ]; then
            docker image build -t $(basename $PWD):test .
        else 
            draft up    
            draft delete 
        fi
    fi
done