#!/bin/bash -ex

export ROOT=${ROOT:-$(git rev-parse --show-toplevel)}
cd "$ROOT/ads-reporting"

./gradlew :caterpie:clean
./gradlew :caterpie:installDist

export WORK_DIR=${WORK_DIR:-$ROOT/applications/caterpie}
cd "$WORK_DIR"