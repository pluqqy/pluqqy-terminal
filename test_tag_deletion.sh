#!/bin/bash
# Test script for verifying tag deletion functionality

echo "Testing Tag Deletion Functionality"
echo "=================================="

# Create a test component with tags
echo "1. Creating test component with tags..."
cat > .pluqqy/components/prompts/test-component.md << 'EOF'
---
tags: [test-tag-1, test-tag-2, shared-tag]
---
This is a test component
EOF

# Create another component with shared tag
echo "2. Creating second component with shared tag..."
cat > .pluqqy/components/prompts/test-component-2.md << 'EOF'
---
tags: [shared-tag, test-tag-3]
---
This is another test component
EOF

# Check initial state
echo "3. Initial components with tags:"
echo "   test-component.md: test-tag-1, test-tag-2, shared-tag"
echo "   test-component-2.md: shared-tag, test-tag-3"

echo ""
echo "4. To test tag deletion:"
echo "   a. Run: ./pluqqy"
echo "   b. Press 'ctrl+t' on test-component.md"
echo "   c. Delete 'test-tag-1' using 'ctrl+d'"
echo "   d. Check that tag is removed from the file"
echo "   e. Try deleting 'shared-tag' - it should warn about usage in other files"

echo ""
echo "5. Check files after deletion:"
echo "   cat .pluqqy/components/prompts/test-component.md"
echo "   cat .pluqqy/components/prompts/test-component-2.md"

echo ""
echo "Test components created. Run ./pluqqy to test the tag deletion feature."