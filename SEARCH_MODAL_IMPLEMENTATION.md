# Integrated Search Implementation

## Overview
An integrated search bar has been implemented for the Pluqqy TUI that allows users to search and filter pipelines and components in real-time without leaving their current view.

## Key Features

### 1. Integrated Search Bar
- **Always visible**: Search bar appears at the top of the main list view
- **Quick access**: Press `/` to jump directly to search from any pane
- **Live filtering**: Results update as you type
- **Context preserved**: No modal overlay, stays in current view
- **Visual indicator**: Uses ⌕ search icon for better visibility

### 2. Search Capabilities
- **Full query syntax support**: 
  - Field searches: `tag:api`, `type:pipeline`, `name:handler`
  - Content search: `content:"error handling"`
  - Date filters: `modified:>7d`
  - Logical operators: `tag:api AND type:prompt`
- **Live filtering**: Both pipeline and component lists filter in real-time
- **Smart matching**: Uses the existing search engine for powerful queries

### 3. Navigation
- **Tab navigation**: Cycles through search → components → pipelines → preview
- **Search shortcuts**:
  - `/`: Jump to search from any pane
  - `Esc`: Clear search and return to components pane
  - Standard text editing within search bar
- **Filtered navigation**:
  - Arrow keys work normally within filtered lists
  - Enter opens/edits selected item as usual
  - All standard shortcuts (e, t, d, etc.) work on filtered items

### 4. Result Display
- Each result shows:
  - Name (with selection indicator when focused)
  - Type (component subtype or "pipeline")
  - Tags (comma-separated list)
- Results counter at the bottom

### 4. Integration
- **Search engine initialization**: Automatically indexes all content on startup
- **Seamless filtering**: 
  - Original lists are preserved
  - Filtered lists update dynamically
  - Cursor positions adjust automatically
- **Consistent UI**: Search bar matches the existing TUI style

## Implementation Details

### Files Modified
1. **pkg/tui/list.go**
   - Added search-related fields: `searchInput`, `searchQuery`, `filteredPipelines`, `filteredComponents`
   - Added `searchPane` to pane navigation
   - Implemented `performSearch()` method for filtering
   - Updated view to render search bar at top
   - Modified component/pipeline rendering to use filtered lists

2. **pkg/tui/app.go**
   - Removed modal-related code
   - Simplified to use integrated search

3. **pkg/search/parser.go**
   - Fixed naming conflict (`FieldType` constant renamed to `FieldTypeField`)

### Architecture Decisions
- **Integrated approach**: Search bar is part of the main UI, not a modal
- **Real-time filtering**: Lists update as you type
- **Minimal disruption**: Users stay in their current workflow
- **Consistent navigation**: All existing shortcuts and navigation work with filtered results

## Usage Examples

### Basic Search
1. Press `/` to jump to search bar
2. Type part of a component or pipeline name
3. Both lists filter in real-time
4. Tab to navigate to the filtered results
5. Use arrow keys and Enter as normal

### Search by Type
1. Type `type:pipeline` to see only pipelines
2. Type `type:prompt` to see only prompt components
3. Clear search to see all items again

### Tag Search
1. Type `tag:api` to see all items with the "api" tag
2. Combine with other filters: `tag:api AND type:component`

### Clear Search
1. Press `Esc` while in search to clear and return to components
2. Or manually delete the search text

## Pipeline Builder Search

The integrated search bar has also been implemented in the pipeline builder view with the following features:

### Search Bar in Pipeline Builder
- **Always visible**: Search bar appears at the top of the builder view
- **Quick access**: Press `/` to jump directly to search from any column
- **Live filtering**: Available components filter in real-time as you type
- **Same query syntax**: Supports all the same search queries as the main view
- **Seamless navigation**: Tab cycles through search → left column → right column → preview
- **Context preserved**: Search doesn't disrupt your pipeline building workflow

### Usage in Pipeline Builder
1. Press `/` from anywhere to jump to search
2. Type to filter available components
3. Tab to navigate to the filtered components
4. Press Enter to add components to your pipeline
5. Esc clears search and returns to left column

## Future Enhancements
- **Search history**: Remember recent searches
- **Search highlighting**: Highlight matching terms in results
- **Fuzzy matching**: Handle typos and partial matches
- **Quick filters**: Buttons for common searches