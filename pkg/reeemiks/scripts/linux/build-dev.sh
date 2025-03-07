#!/bin/sh

echo 'Building reeemiks (development)...'

# shove git commit, version tag into env
GIT_COMMIT=$(git rev-list -1 --abbrev-commit HEAD)
VERSION_TAG=$(git describe --tags --always)
BUILD_TYPE=dev
echo 'Embedding build-time parameters:'
echo "- gitCommit $GIT_COMMIT"
echo "- versionTag $VERSION_TAG"
echo "- buildType $BUILD_TYPE"

go build -o reeemiks-dev -ldflags "-X main.gitCommit=$GIT_COMMIT -X main.versionTag=$VERSION_TAG -X main.buildType=$BUILD_TYPE" ./pkg/reeemiks/cmd
if [ $? -eq 0 ]; then
    echo 'Done.'
else
    echo 'Error: "go build" exited with a non-zero code.'
    exit 1
fi

