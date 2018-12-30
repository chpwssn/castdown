#!/usr/bin/env bash

package_name="castdown"

platforms=("windows/amd64" "windows/386" "darwin/amd64" "freebsd/386" "freebsd/amd64" "freebsd/arm" "linux/arm" "linux/amd64" "linux/386" "linux/arm64" "openbsd/386" "openbsd/amd64" "openbsd/arm")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi  

    mkdir -p dist
    env GOOS=$GOOS GOARCH=$GOARCH go build -o dist/$output_name $package
    if [ $? -ne 0 ]; then
        echo "Building failed for $GOOS $GOARCH"
        exit 1
    fi
done
