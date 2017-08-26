#!/bin/sh

# Writes the current tag to standard out. If the current commit isn't
# tagged then the current branch name is used.
tag=$(git describe --tags --exact-match 2>/dev/null)
if [ "$?" != 0 ]
then
    tag=$(git rev-parse --abbrev-ref HEAD)
fi

echo $tag
