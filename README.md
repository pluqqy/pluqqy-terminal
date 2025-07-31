# Pluqqy

Build and manage LLM prompt pipelines from your terminal.

Pluqqy lets you create reusable components (contexts, prompts, and rules) and combine them into pipelines. When you set a pipeline, it generates a PLUQQY.md file that contains all your composed instructions.

In Claude Code or other AI coding tools, simply reference @PLUQQY.md instead of copying and pasting prompts. Need different instructions? Set a different pipeline - the file updates automatically, but you keep referencing the same @PLUQQY.md.

This approach keeps your context minimal and focused - only including what's relevant for the current task. Both you and the AI work from the same single source of truth, eliminating confusion about which instructions are active while preserving valuable context window space.

## Features

- 🎨 **Beautiful TUI** - Clean, intuitive terminal interface with Pluqqy ASCII art branding
- 📊 **Token Counter** - Real-time token estimation for your composed pipelines
- 🔍 **Component Grouping** - Organized view of components by type with metadata
- 📝 **Built-in Editor** - Edit components directly in the TUI without external editors
- 🔄 **Live Preview** - See your composed pipeline in real-time as you build
- ⚡ **Duplicate Prevention** - Automatically prevents adding the same component twice
- 💾 **Smart Save** - Confirmation prompts to prevent accidental overwrites
- 🛡️ **Exit Protection** - Confirmation dialogs prevent losing unsaved changes

## Installation

```bash
git clone https://github.com/pluqqy/pluqqy-cli
cd pluqqy-cli
make install
```

This will build and install `pluqqy` to `$GOPATH/bin` or `$HOME/go/bin` (if GOPATH is not set).

### Updating

To update to the latest version:

```bash
cd pluqqy-cli
make update
```

Or manually:

```bash
git pull
make install
```

## Usage

### Initialize a new project

```bash
pluqqy init
```

This creates the following structure:

```
.pluqqy/
├── pipelines/
└── components/
    ├── contexts/
    ├── prompts/
    └── rules/
```

### Launch the TUI

```bash
pluqqy
```

### TUI Commands

#### Pipeline List View

- `↑/↓` or `j/k` - Navigate pipelines
- `Enter` - View pipeline details
- `e` - Edit selected pipeline
- `n` - Create new pipeline
- `d` - Delete pipeline (with confirmation)
- `S` - Set selected pipeline (generates PLUQQY.md)
- `r` - Refresh pipeline list
- `q` - Quit
- `Ctrl+C` - Quit (double Ctrl+C to confirm)

#### Pipeline Builder

- `Tab` - Switch between panes (components, selected, preview)
- `↑/↓` - Navigate items
- `Enter` - Add/edit component
- `n` - Create new component
- `e` - Edit component in TUI
- `E` - Edit component with external editor
- `Del` - Remove selected component
- `K/J` - Reorder selected components (move up/down)
- `p` - Toggle preview
- `Ctrl+S` - Save pipeline
- `S` - Save and set as active pipeline
- `Esc` - Back to pipeline list

#### Pipeline Viewer

- `Tab` - Switch between components and preview panes
- `↑/↓` - Scroll in active pane
- `S` - Set as active pipeline
- `e` - Edit pipeline
- `Esc` - Back to pipeline list
- `Ctrl+C` - Quit

#### Component Editor

- Type content directly in the TUI
- `Ctrl+S` - Save component (with overwrite confirmation if needed)
- `Esc` - Cancel

## UI Features

- **Token Counter** - Shows estimated token count in the preview pane with color-coded status (green/yellow/red)
- **Component Metadata** - View file sizes and paths in the component list
- **Scrollable Panes** - All panes support smooth scrolling for long content
- **Help Footer** - Context-sensitive help text at the bottom of each screen
- **Status Messages** - Clear feedback for all actions with auto-dismiss
- **Responsive Layout** - Adapts to terminal size with proper content wrapping

## Output

The `set` command generates a `PLUQQY.md` file in your project root with sections:

- `CONTEXT` - Combined context components
- `PROMPTS` - Combined prompt components  
- `IMPORTANT RULES` - Combined rules components

## Example

1. Initialize project: `pluqqy init`
2. Launch TUI: `pluqqy`
3. Press `n` to create new pipeline
4. Name it "my-assistant"
5. Add components using the builder
6. Press `s` to save
7. Back at pipeline list, press `s` to set the pipeline
8. Check `PLUQQY.md` for the composed output

## Requirements

- Go 1.19 or higher (for building)
- Terminal with UTF-8 support

## License

MIT
