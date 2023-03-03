#!/bin/sh -e

# Save current directory
current=$(dirname $(realpath $0))

# Create temporary directory
repo=$(mktemp -d)
trap "rm -rf ${repo}" EXIT

# Checkout MLFlow source
git clone --depth 1 --branch $(cat ${current}/version) https://github.com/aimhubio/aim.git ${repo}

# Apply our customizations
cd ${repo}
git apply -p1 <${current}/custom.patch

# Build the UI
cd aim/web/ui
npm install
npm run build

# Move the built UI to its destination
[ -d ${current}/build.previous ] && rm -rf ${current}/build.previous
[ -d ${current}/build ] && mv ${current}/build ${current}/build.previous
mv build ${current}/build
