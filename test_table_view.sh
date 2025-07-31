#!/bin/bash
# Test script for component table view

echo "Testing Pluqqy Component Table View"
echo "=================================="
echo ""
echo "To test the new table view:"
echo "1. Run: go run cmd/pluqqy/main.go"
echo "2. Select 'Build New Pipeline' or press 'b'"
echo "3. Enter a pipeline name"
echo "4. Look at the left column - you should see a table with:"
echo "   - Component names"
echo "   - Last modified timestamps"
echo "   - Usage counts with visual bars"
echo ""
echo "Existing functionality to verify:"
echo "- Tab: Switch between columns"
echo "- Enter: Add components (should still work)"
echo "- E: Edit in external editor"
echo "- e: Edit in TUI editor"
echo "- Components with ✓ are already in pipeline"
echo "- All keyboard shortcuts should work as before"
echo ""
echo "Press any key to start testing..."
read -n 1
go run cmd/pluqqy/main.go