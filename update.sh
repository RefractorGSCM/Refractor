#!/bin/bash

echo "This script will checkout the latest tagged version from both the Refractor and Refractor-Svelte repositories."

echo ""
echo "Checking out latest tag of backend service..."
git fetch --tags
latestTag=$(git describe --tags `git rev-list --tags --max-count=1`)

git checkout $latestTag  > /dev/null
echo ""
echo "Successfully checked out latest tag."

echo ""
echo "Checking out latest tag of frontend application..."
cd Refractor-Svelte
git fetch --tags
latestTag=$(git describe --tags `git rev-list --tags --max-count=1`)

git checkout $latestTag > /dev/null
cd ..
echo ""
echo "Successfully checked out latest tag."

echo ""
echo ""
echo "Rebuilding services..."
echo "Please let this script run uninterrupted until it's finished."

docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml up -d --force-recreate --build --no-deps refractor refractor-frontend

echo ""
echo ""
echo "Services updated!"
