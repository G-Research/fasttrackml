#!/bin/sh -e

# Initialize variables
current=$(dirname $(realpath $0))
workspace=$(realpath ${current}/../../../..)
version=$(cat ${current}/version)
repo=${current}/mlflow.src
venv=/tmp/venv-$(echo ${repo} | sha256sum | awk '{print $1}')

# Reset repo if needed
if [ ${current}/version -nt ${repo} ] || [ ${current}/custom.patch -nt ${repo} ]
then
  rm -rf ${repo}
fi

# Download and patch repo if needed
if [ ! -d ${repo} ]
then
  # Checkout MLFlow source
  git clone --depth 1 --branch ${version} https://github.com/mlflow/mlflow.git ${repo}

  # Apply our customizations
  cd ${repo}
  git apply -p1 <${current}/custom.patch
fi

cd ${repo}

# Create venv if needed
if [ ! -d ${venv} ]
then
  python -mvenv ${venv}
  . ${venv}/bin/activate
  python setup.py -q dependencies | pip install -r requirements/test-requirements.txt -r /dev/stdin
  deactivate
fi

# Build fml
make -C ${workspace} build
cp ${workspace}/fml ${repo}/fml

# Create postgres test database if needed
psql postgres://postgres:postgres@localhost <<EOF
SELECT 'CREATE DATABASE test'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'test')\gexec
EOF

# Run tests
. ${venv}/bin/activate
export PATH=".:${PATH}"
pytest tests/tracking/test_rest_tracking.py
