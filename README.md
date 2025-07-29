# Pluqqy

Build and manage LLM prompt pipelines from your terminal.

Pluqqy lets you create reusable components (contexts, prompts, and rules) and combine them into pipelines. When you set a pipeline, it generates a PLUQQY.md file that contains all your composed instructions.

In Claude Code or other AI coding tools, simply reference @PLUQQY.md instead of copying and pasting prompts. Need different instructions? Set a different pipeline - the file updates automatically, but you keep referencing the same @PLUQQY.md.

This approach keeps your context minimal and focused - only including what's relevant for the current task. Both you and the AI work from the same single source of truth, eliminating confusion about which instructions are active while preserving valuable context window space.

## Installation

```bash
git clone https://github.com/pluqqy/pluqqy-cli
cd pluqqy-cli
make build
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

- `↑/↓` - Navigate pipelines
- `Enter` - View pipeline details
- `n` - Create new pipeline
- `e` - Edit selected pipeline
- `s` - Set selected pipeline (generates PLUQQY.md)
- `q` - Quit

#### Pipeline Builder

- `Tab` - Switch between component list and selected components
- `↑/↓` - Navigate items
- `Enter` - Add/remove component
- `Ctrl+↑/↓` - Reorder selected components
- `n` - Create new component
- `e` - Edit component
- `p` - Toggle preview
- `s` - Save pipeline
- `Esc` - Back to pipeline list

#### Component Editor

- Type content directly in the TUI
- `Ctrl+S` - Save component
- `Esc` - Cancel

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
