#!/bin/sh -e

# current directory and checkout location
current=$(dirname $(realpath $0))
repo="${current}/repo"

# Checkout source and build if necessary
if [ ! -d "${repo}" ]; then
  git clone --depth 1 -b $(cat ${current}/version) https://github.com/aimhubio/aim.git ${repo}

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
fi
