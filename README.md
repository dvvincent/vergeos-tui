# vergeos-tui

A terminal UI for managing [VergeOS](https://www.verge.io), driving the `vrg`
CLI in the background. Built on [Bubble Tea](https://github.com/charmbracelet/bubbletea).

> Point-and-manage your VergeOS cloud from the terminal — no browser, no
> memorizing subcommands. `vergeos-tui` shells out to `vrg` for every
> operation, so it inherits whatever profile/credentials `vrg` is already
> configured with.

![vergeos-tui demo](demo.gif)

## Status

**v0.3** — tabbed resource browser with the polish:

- **Tabs:** VMs · Networks · Tenants · Storage (`tab` / `⇧tab`, lazy-loaded)
- **Cluster-context header** — host, cloud, VMs online/total, nodes, alarms, version
- **Colored status** — green running · red stopped · yellow transitional; storage
  `USED%` heat-colored (green/yellow/red)
- **Search** (`/`) — live substring filter with match highlighting
- **Sort** — number keys sort by that column; press again to reverse, again to clear
  (`▲`/`▼` shown in the header)
- **VM detail** (`enter`): summary + drives + NICs
- **VM lifecycle actions:** start (`s`) · stop (`x`) · restart (`R`), behind a confirm modal
- **Auto-refresh** (`a`): polls the active tab (and the header) every 5s
- Page nav (`pgup`/`pgdn`, `g`/`G`), refresh (`r`), contextual help (`?`)

Rendered with a small custom ANSI-aware table (bubbles/table can't colorize cells —
it truncates with runewidth). Logic is unit-tested (`go test ./...`).

Idioms borrowed from the [Bubble Tea examples](https://github.com/charmbracelet/bubbletea/tree/main/examples):
the **notebook-style tabs** (connected active-tab border) follow `examples/tabs`, and
search uses the **`textinput`** bubble (`examples/textinput`).

## Requirements

> **`vrg` must be installed and already configured — this is a hard
> prerequisite.** The TUI has no VergeOS endpoint, credentials, or config of
> its own; it shells out to `vrg` for every operation and inherits whatever
> profile `vrg` is pointed at. **Acceptance test:** if `vrg system info`
> returns your cloud, the TUI works — if it errors, fix `vrg` first.

- [`vrg`](https://docs.verge.io) installed and configured (verify with `vrg system info`)
- Go 1.24+ to build

## Build & run

```bash
go build -o vtui .
./vtui                 # launch the TUI
```

Non-interactive helpers (no TTY needed):

```bash
./vtui --selftest      # exercise the vrg backend, print a few VMs
./vtui --preview       # render one static frame to stdout
```

## Keys

| Key             | Action                             |
|-----------------|------------------------------------|
| `↑`/`k` `↓`/`j` | move                               |
| `pgup`/`pgdn`   | page up / down                     |
| `g` / `G`       | top / bottom                       |
| `tab` / `⇧tab`  | next / previous tab                |
| `/`             | search (live filter)               |
| `1`–`7`         | sort by column (again = reverse)   |
| `enter`         | detail view (any tab)              |
| `↑`/`↓` in detail | scroll long detail panels        |
| `s` / `x` / `R` | start / stop / restart VM          |
| `r`             | refresh active tab                 |
| `a`             | toggle auto-refresh (5s countdown) |
| `?`             | toggle help                        |
| `^Z`            | suspend (`fg` to resume)           |
| `esc`           | back (from detail/confirm/search)  |
| `q` / `^C`      | quit                               |

**Mouse:** click a tab to switch, click a row to select (click again to open
its detail), and use the wheel to scroll the list or a detail panel.

Preview any tab non-interactively (renders one static frame):

```bash
./vtui --preview            # VMs
./vtui --preview networks   # or tenants | storage | detail
```

Regenerate the demo GIF (needs [vhs](https://github.com/charmbracelet/vhs) + ffmpeg + ttyd):

```bash
vhs demo.tape                       # → demo.gif
VHS_NO_SANDBOX=1 vhs demo.tape      # in containers/CI (no Chrome sandbox)
```

The tape is **read-only** — it navigates, searches, sorts, and opens detail;
it never triggers start/stop, so it won't mutate a cluster.

## Architecture

Elm architecture (Model → Update → View), the Bubble Tea way:

```
main.go                 entry point; --selftest / --preview helpers
internal/vrg/           backend: exec `vrg -q -o json ...`, parse into structs
internal/ui/
  model.go              app model: tabs, modes (list/detail/confirm), commands
  keys.go               key bindings + help
  detail.go             VM detail + confirm-modal rendering
  styles.go             lipgloss styles
```

Every `vrg` call is a `tea.Cmd` that runs off the UI goroutine and returns a
`Msg`, so the interface never blocks while the CLI is working. Tabs are
**lazy-loaded** on first visit. The `vrg` package is UI-free and independently
testable (that's what `--selftest` drives — it exercises every backend + the
detail path).

## Roadmap

- Snapshots tab + snapshot create/restore
- Network/tenant actions (not just detail views)
- Profile switcher (`vrg -p <profile>`) for multi-cloud
- Responsive column widths on narrow terminals

## License

MIT — see [LICENSE](LICENSE).
