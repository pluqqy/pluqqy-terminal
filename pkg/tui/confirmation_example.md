# Confirmation Module Usage Example

## Before (current implementation in list.go):

```go
// In struct:
confirmingDelete   bool
deleteConfirmation string
deletingFromPane   pane

// In Update method:
case "d", "delete":
    if m.activePane == pipelinesPane {
        if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
            m.confirmingDelete = true
            m.deletingFromPane = pipelinesPane
            m.deleteConfirmation = fmt.Sprintf("Delete pipeline '%s'? %s", 
                m.pipelines[m.pipelineCursor].name, formatConfirmOptions(true))
        }
    }

// In key handling:
if m.confirmingDelete {
    switch msg.String() {
    case "y", "Y":
        m.confirmingDelete = false
        // ... execute delete
    case "n", "N", "esc":
        m.confirmingDelete = false
        m.deleteConfirmation = ""
    }
}

// In View:
if m.confirmingDelete {
    confirmStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("196")).
        MarginTop(2)
    s.WriteString(contentStyle.Render(confirmStyle.Render(m.deleteConfirmation)))
}
```

## After (using confirmation module):

```go
// In struct:
deleteConfirm      *ConfirmationModel
deletingFromPane   pane

// In NewMainListModel:
m.deleteConfirm = NewConfirmation()

// In Update method:
case "d", "delete":
    if m.activePane == pipelinesPane {
        if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
            m.deletingFromPane = pipelinesPane
            pipelineName := m.pipelines[m.pipelineCursor].name
            
            m.deleteConfirm.ShowInline(
                fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
                true, // destructive
                func() tea.Cmd {
                    return m.deletePipeline(pipelineName)
                },
                nil, // onCancel
            )
        }
    }

// In key handling:
if m.deleteConfirm.Active() {
    return m, m.deleteConfirm.Update(msg)
}

// In View:
if m.deleteConfirm.Active() {
    confirmStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("196")).
        MarginTop(2)
    s.WriteString(contentStyle.Render(confirmStyle.Render(m.deleteConfirm.View())))
}
```

## For Dialog-style confirmations (exit confirmation):

```go
// In struct:
exitConfirm *ConfirmationModel

// Show exit confirmation:
m.exitConfirm.ShowDialog(
    "⚠️  Unsaved Changes",
    "You have unsaved changes in this component.",
    "Exit without saving?",
    true, // destructive
    m.width - 4,
    10,
    func() tea.Cmd {
        // Exit logic
        m.editingComponent = false
        return nil
    },
    nil, // onCancel - just hide the dialog
)

// In Update:
if m.exitConfirm.Active() {
    return m, m.exitConfirm.Update(msg)
}

// In View (replaces entire exitConfirmationView method):
if m.exitConfirm.Active() {
    return m.exitConfirm.View()
}
```

## Benefits:
1. **Less code duplication** - No need to track confirmation state separately
2. **Consistent behavior** - All confirmations work the same way
3. **Easier to maintain** - Changes to confirmation logic in one place
4. **Type safety** - Configuration is validated at compile time
5. **Flexible** - Easy to add new confirmation types or styles