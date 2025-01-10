#!/bin/bash

# Check if a tag name was provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <tag_name>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

TAG_NAME=$1

# Validate tag format (should start with 'v')
if [[ ! $TAG_NAME =~ ^v ]]; then
    echo "Error: Tag name must start with 'v' (e.g., v1.0.0)"
    exit 1
fi

# Confirm with user
echo "This will:"
echo "1. Clean and rebuild all packages"
echo "2. Force-update tag '$TAG_NAME' to point to the current commit"
echo "3. Push the tag to trigger the GitHub action"
echo
echo "This is destructive and will overwrite the remote tag."
read -p "Are you sure you want to continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operation cancelled"
    exit 1
fi

# Ensure .env exists
if [ ! -f .env ]; then
    echo "Creating empty .env file..."
    touch .env
fi

# Clean and rebuild
echo "Cleaning and rebuilding packages..."
make clean
if ! make dist-package-all; then
    echo "Error: Build failed"
    exit 1
fi

# Force update the tag locally
echo "Updating local tag..."
git tag -f $TAG_NAME

# Force push to remote
echo "Force pushing tag to remote..."
git push -f origin $TAG_NAME

echo "Tag '$TAG_NAME' has been updated and force-pushed to remote."
echo "The GitHub action should trigger automatically." 