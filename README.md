# Pluqqy

A terminal UI for building modular LLM prompt pipelines. Create reusable components (contexts, prompts, rules) and compose them into powerful prompt templates.

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
