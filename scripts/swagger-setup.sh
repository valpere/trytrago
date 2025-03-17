#!/bin/bash

# This script will download and prepare the Swagger UI files
# Make this script executable with: chmod +x setup.sh

# set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

TARGET_DIR="${SCRIPT_DIR}/../interface/api/rest/docs"

pushd "${TARGET_DIR}"

# Create necessary directories
mkdir -p "swagger-ui"

# Download the latest Swagger UI release
curl -L https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.9.0.tar.gz -o swagger-ui.tar.gz

# Extract only the dist directory
tar -xzf swagger-ui.tar.gz --strip-components=2 swagger-ui-5.9.0/dist

# Move files to swagger-ui directory
mv *.js *.css *.html *.png favicon-* swagger-ui/

# Copy our OpenAPI specification
cp ../../../../docs/OpenAPISpecification.yaml openapi.yaml

# Clean up
rm swagger-ui.tar.gz

popd

echo "Swagger UI setup complete!"
