#!/usr/bin/env bash

set -eE
set -o pipefail

# This is a wrapper around load.py to faciliate exporting the environment
# variables in the current shell

# E.g., . load.sh --file file1 --file file2 --file file3 \
#                 --override override1 --override override2 --override override3

ROOT=$(git rev-parse --show-toplevel)
WDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

tmpfile=$(mktemp)

declare -a files overrides

while [[ $# -gt 0 ]]; do
    case "$1" in
        --file) files+=($2); shift 2;;
        --override) overrides+=($2); shift 2;;
        *) echo "Invalid parameter $1" && exit 1
    esac
done

$WDIR/load.py -f ${files[@]} -o ${overrides[@]} > $tmpfile

cat $tmpfile
. $tmpfile
rm $tmpfile
