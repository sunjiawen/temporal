#!/usr/bin/env bash

set -xeE
set -o pipefail

ROOT=$(git rev-parse --show-toplevel)

$ROOT/config/test/test_config.sh
