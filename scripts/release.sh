#!/bin/bash

# Pluqqy Release Build Script
# This script creates release builds for all supported platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from user or use default
VERSION=${1:-"v0.1.0"}
RELEASE_DIR="releases/${VERSION}"

echo -e "${GREEN}Building Pluqqy ${VERSION}${NC}"

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf releases/${VERSION}
mkdir -p ${RELEASE_DIR}

# Run tests first
echo -e "${YELLOW}Running tests...${NC}"
make test

# Build for all platforms
echo -e "${YELLOW}Building for all platforms...${NC}"
make build-all

# Create release archives
echo -e "${YELLOW}Creating release archives...${NC}"

# Function to create archive
create_archive() {
    local platform=$1
    local binary=$2
    local archive_name="${RELEASE_DIR}/pluqqy-${VERSION}-${platform}"
    
    echo "Creating archive for ${platform}..."
    
    # Create temporary directory
    local temp_dir="temp_${platform}"
    mkdir -p ${temp_dir}
    
    # Copy binary and examples
    cp build/${binary} ${temp_dir}/
    cp README.md ${temp_dir}/
    cp -r examples ${temp_dir}/
    
    # Create appropriate archive
    if [[ ${platform} == *"windows"* ]]; then
        # Create zip for Windows
        cd ${temp_dir}
        zip -r "../${archive_name}.zip" .
        cd ..
    else
        # Create tar.gz for Unix systems
        cd ${temp_dir}
        tar -czf "../${archive_name}.tar.gz" .
        cd ..
    fi
    
    # Clean up
    rm -rf ${temp_dir}
}

# Create archives for each platform
create_archive "darwin-amd64" "pluqqy-darwin-amd64"
create_archive "darwin-arm64" "pluqqy-darwin-arm64"
create_archive "linux-amd64" "pluqqy-linux-amd64"
create_archive "linux-arm64" "pluqqy-linux-arm64"
create_archive "windows-amd64" "pluqqy-windows-amd64.exe"

# Generate checksums
echo -e "${YELLOW}Generating checksums...${NC}"
cd ${RELEASE_DIR}
shasum -a 256 *.tar.gz *.zip > checksums.txt
cd ../..

# Create release notes template
echo -e "${YELLOW}Creating release notes template...${NC}"
cat > ${RELEASE_DIR}/RELEASE_NOTES.md << EOF
# Pluqqy ${VERSION} Release Notes

## What's New
- Initial release of Pluqqy
- Terminal-based UI for managing LLM prompt pipelines
- Support for creating, editing, and composing pipelines
- Cross-platform support (macOS, Linux, Windows)

## Installation

### macOS (Intel)
\`\`\`bash
curl -L https://github.com/yourusername/pluqqy/releases/download/${VERSION}/pluqqy-${VERSION}-darwin-amd64.tar.gz | tar xz
chmod +x pluqqy
sudo mv pluqqy /usr/local/bin/
\`\`\`

### macOS (Apple Silicon)
\`\`\`bash
curl -L https://github.com/yourusername/pluqqy/releases/download/${VERSION}/pluqqy-${VERSION}-darwin-arm64.tar.gz | tar xz
chmod +x pluqqy
sudo mv pluqqy /usr/local/bin/
\`\`\`

### Linux (amd64)
\`\`\`bash
curl -L https://github.com/yourusername/pluqqy/releases/download/${VERSION}/pluqqy-${VERSION}-linux-amd64.tar.gz | tar xz
chmod +x pluqqy
sudo mv pluqqy /usr/local/bin/
\`\`\`

### Linux (arm64)
\`\`\`bash
curl -L https://github.com/yourusername/pluqqy/releases/download/${VERSION}/pluqqy-${VERSION}-linux-arm64.tar.gz | tar xz
chmod +x pluqqy
sudo mv pluqqy /usr/local/bin/
\`\`\`

### Windows
1. Download \`pluqqy-${VERSION}-windows-amd64.zip\`
2. Extract the archive
3. Add the directory to your PATH

## Checksums
See \`checksums.txt\` for SHA-256 checksums of all release files.
EOF

# Summary
echo -e "${GREEN}✓ Release ${VERSION} built successfully!${NC}"
echo -e "${GREEN}✓ Release files created in: ${RELEASE_DIR}${NC}"
echo ""
echo "Files created:"
ls -la ${RELEASE_DIR}/
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Review the release notes in ${RELEASE_DIR}/RELEASE_NOTES.md"
echo "2. Test the binaries on each platform"
echo "3. Create a GitHub release and upload the archives"
echo "4. Tag the release: git tag ${VERSION} && git push origin ${VERSION}"