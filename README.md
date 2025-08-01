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
├── components/
│   ├── contexts/
│   ├── prompts/
│   └── rules/
├── archive/
│   ├── pipelines/
│   └── components/
│       ├── contexts/
│       ├── prompts/
│       └── rules/
├── tmp/              # For pipeline-generated output files
└── .gitignore        # Ignores tmp directory
```

### Launch the TUI

```bash
pluqqy
```

### TUI Commands

#### Main List View

- `Tab` - Switch between pipelines and components panes
- `↑/↓` or `j/k` - Navigate items
- `Enter` - View/edit pipeline or component
- `e` - Edit component in TUI / Edit pipeline in builder
- `E` - Edit component with external editor (components pane only)
- `n` - Create new pipeline/component
- `a` - Archive pipeline/component (with confirmation)
- `d` - Delete pipeline/component (with confirmation)
- `S` - Set selected pipeline (generates PLUQQY.md)
- `s` - Open settings editor
- `p` - Toggle preview pane
- `r` - Refresh pipeline list
- `Ctrl+C` - Quit (double Ctrl+C to confirm)

#### Pipeline Builder

- `Tab` - Switch between panes (available components, pipeline components, preview)
- `↑/↓` - Navigate items
- `Enter` - Add/remove component (toggles)
- `n` - Create new component
- `e` - Edit component in TUI
- `E` - Edit component with external editor
- `Del`, `d`, `Backspace` - Remove component from pipeline (right pane)
- `K/J` or `Ctrl+↑/↓` - Reorder pipeline components (move up/down)
- `p` - Toggle preview pane
- `Ctrl+S` - Save pipeline
- `S` - Save and set as active pipeline
- `Esc` - Back to main list (with unsaved changes confirmation)

#### Pipeline Viewer

- `Tab` - Switch between components and preview panes
- `↑/↓` - Scroll in active pane
- `S` - Set as active pipeline
- `e` - Edit pipeline
- `Esc` - Back to pipeline list
- `Ctrl+C` - Quit

#### Component Editor (TUI)

- Type content directly in the editor
- `↑/↓` - Scroll through content
- `Ctrl+S` - Save component
- `E` - Open in external editor
- `Esc` - Cancel (with unsaved changes confirmation)

## UI Features

- **Token Counter** - Shows estimated token count in the preview pane with color-coded status (green/yellow/red)
- **Component Metadata** - View file sizes and paths in the component list
- **Scrollable Panes** - All panes support smooth scrolling for long content
- **Help Footer** - Context-sensitive help text at the bottom of each screen
- **Status Messages** - Clear feedback for all actions with auto-dismiss
- **Responsive Layout** - Adapts to terminal size with proper content wrapping

## Output

The `set` command generates a `PLUQQY.md` file in your project root. This naming convention is intentional:

- **Non-conflicting**: Won't overwrite common files like `AGENT.md` or `CLAUDE.md`
- **Claude Code Integration**: Simply reference `@PLUQQY.md` in Claude Code to load your entire pipeline
- **Pipeline Agnostic**: No need to remember specific pipeline names
- **Chainable**: Easily reference and combine multiple pipelines in Claude Code sessions
- **Customizable**: Change the default filename in settings (press `s` from main view)

The file contains sections in your configured order:

- `CONTEXT` - Combined context components
- `PROMPTS` - Combined prompt components
- `IMPORTANT RULES` - Combined rules components

**Tip for teams**: To keep `PLUQQY.md` tracked in git but ignore local changes:
```bash
git update-index --skip-worktree PLUQQY.md
```
This lets each developer use different pipelines without creating commit noise.

## Example

1. Initialize project: `pluqqy init`
2. Launch TUI: `pluqqy`
3. Press `n` to create new pipeline
4. Name it "my-assistant"
5. Add components using the builder
6. Press `s` to save
7. Back at pipeline list, press `s` to set the pipeline
8. Check `PLUQQY.md` for the composed output

## Configuration

### Settings Editor

Pluqqy includes a built-in settings editor accessible from the TUI. Press `s` from the main list view to customize:

- **Output Settings**
  - Default filename for generated output (default: `PLUQQY.md`)
  - Export path for pipeline output files (default: `./` - your project root)
  - Output path for pipeline-generated files (default: `.pluqqy/tmp/`)

- **Formatting Options**
  - Toggle section headings in output
  - Reorder sections using `J/K` keys
  - Edit section types and headings

Changes take effect immediately upon saving with `S`.

### External Editor

Pluqqy uses your system's `$EDITOR` environment variable to determine which external editor to use. Set it in your shell configuration:

```bash
# For vim users
export EDITOR=vim

# For nano users
export EDITOR=nano

# For VS Code users
export EDITOR="code --wait"

# For Cursor users
export EDITOR="cursor --wait"

# For Windsurf users
export EDITOR="windsurf --wait"

# For Zed users
export EDITOR="zed --wait"
```

## Requirements

- Go 1.19 or higher (for building)
- Terminal with UTF-8 support

## License

MIT
