#!/usr/bin/env bash

package=$1
package_name=$2

if [[ -z "$package" || -z "$package_name" ]]; then
  echo "usage: ./build.sh <package-path> <name>"
  echo "e.g.: ./build/build.sh ./cmd/toolname/ toolname"
  exit 1
fi

platforms=("windows/amd64" "linux/amd64" "darwin/amd64" "darwin/arm64")

# Make sure build directory exists
mkdir -p build

for platform in "${platforms[@]}"
do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    output_name=$package_name'-'$GOOS'-'$GOARCH

    if [ "$GOOS" = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "build/$output_name" $package

    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

ls -la build
echo "To compress even more, use 'upx --brute <file>'"