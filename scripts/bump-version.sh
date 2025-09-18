#!/bin/bash

# Script to manually bump version and create a release tag

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo -e "${GREEN}Current version: ${LATEST_TAG}${NC}"

# Parse current version
VERSION=${LATEST_TAG#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Ask what type of release
echo ""
echo "What type of release is this?"
echo "1) Patch (bug fixes)      - $MAJOR.$MINOR.$((PATCH + 1))"
echo "2) Minor (new features)   - $MAJOR.$((MINOR + 1)).0"
echo "3) Major (breaking changes) - $((MAJOR + 1)).0.0"
echo "4) Custom version"
echo ""

read -p "Select release type (1-4): " RELEASE_TYPE

case $RELEASE_TYPE in
    1)
        NEW_VERSION="v$MAJOR.$MINOR.$((PATCH + 1))"
        ;;
    2)
        NEW_VERSION="v$MAJOR.$((MINOR + 1)).0"
        ;;
    3)
        NEW_VERSION="v$((MAJOR + 1)).0.0"
        ;;
    4)
        read -p "Enter custom version (e.g., 1.2.3): " CUSTOM_VERSION
        # Remove 'v' if user included it
        CUSTOM_VERSION=${CUSTOM_VERSION#v}
        NEW_VERSION="v$CUSTOM_VERSION"
        ;;
    *)
        echo -e "${RED}Invalid selection${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${YELLOW}New version will be: ${NEW_VERSION}${NC}"
echo ""

# Get release notes
echo "Enter release notes (press Ctrl+D when done):"
NOTES=$(cat)

# Confirm
echo ""
read -p "Create release ${NEW_VERSION}? (y/n) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Create and push tag
    git tag -a "$NEW_VERSION" -m "$NOTES"
    git push origin "$NEW_VERSION"
    
    echo ""
    echo -e "${GREEN}✅ Release ${NEW_VERSION} created!${NC}"
    echo ""
    echo "The GitHub Actions workflow will now:"
    echo "- Build binaries for all platforms"
    echo "- Create a GitHub release"
    echo "- Publish artifacts"
else
    echo -e "${RED}Release cancelled${NC}"
fi
