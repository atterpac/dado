---
label: Guides
icon: book
order: 70
expanded: true
---

# Guides

In-depth documentation for Jig's core systems.

---

## Available Guides

| Guide | Description |
|-------|-------------|
| [Architecture](architecture.md) | Threading model, lifecycle, and design patterns |
| [Theming](theming.md) | Theme system, built-in themes, custom themes |
| [Navigation](navigation.md) | Page stack, modals, focus management |
| [Data Binding](data-binding.md) | Observable values, computed values, form binding |
| [Input Handling](input-handling.md) | KeyBindings builder, Vim navigation, actions |
| [Troubleshooting](troubleshooting.md) | Common issues and solutions |

---

## Quick Reference

### Design Principles

1. **Composition over Inheritance** - Components wrap primitives, not extend them
2. **Explicit Lifecycle** - Start/Stop methods make state transitions predictable
3. **Lock-Free Theme Reads** - Atomic storage prevents Draw() deadlocks
4. **Progressive Disclosure** - Simple APIs for common cases, power features when needed

### Key Rules

!!!danger Critical
All UI mutations must happen on the main goroutine. Use `QueueUpdateDraw` for async updates.
!!!

!!!warning Important
Always unsubscribe from bindings and unregister from the theme system when components are destroyed.
!!!

!!!success Best Practice
Read theme colors at draw time, not at creation time, to support runtime theme switching.
!!!
