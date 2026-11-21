# Writing user documentation

For people who use conm and have never seen its source. `README.md` is the entry point;
`docs/user/` is the sequential guide behind it, indexed by `docs/user/README.md`.

## The pages

| Page | Covers |
|---|---|
| `first-run.md` | what `conm` shows before any connection exists |
| `add-connection.md` | the add form, prefilling it from a command line or URL, importing from a client's config file and exporting back to it |
| `edit-connection.md` | editing and removing, from the table and from the shell |
| `connect-and-execute.md` | opening the client, running a query or a script |
| `secrets.md` | the secrets screen, keyring references, secrets per workspace |
| `ssh-transport.md` | reaching a database through an SSH transport |
| `workspaces.md` | separate sets of connections and secrets |
| `additional-screens.md` | settings (clients, keybindings) and the info screen |
| `commands.md` | every shell command in short, and `conm doctor` |

## When a feature lands

1. Find the page that owns the behaviour and change it there. Extend a page before adding
   one; a new page is for a new *area*, not a new option.
2. A new or changed shell command also changes `commands.md`.
3. A new page is added to `docs/user/README.md` — its numbered entry and its section links —
   and to the chain: every page ends with a `[← previous] · [Contents](README.md) · [next →]`
   line, so the new page gets one and its two neighbours' lines are re-pointed.
4. A changed section heading changes every anchor that points at it (`README.md`,
   `docs/user/README.md`, other pages).
5. `README.md` changes only when installation, the feature list or the first steps change.
6. Record or re-record the GIF that shows the behaviour (below).

## How a page is written

- **Say what to do, not which key does it.** Every screen lists its own keys behind `?`, and
  keys can be rebound; a page that names a key goes stale. Name the action — "open the
  secrets screen from the connections table".
- **Postgres in every example**, the `orders` database and names like `orders-local`; say
  once where another type differs instead of repeating the example per type.
- **Task first.** A page opens with what the reader gets, then the steps, in the order they
  are performed. "From the connections table" and "From the shell" are the two usual
  sections when a task can be done both ways.
- Shell commands in `sh` fences, exactly as typed; file contents in `toml` fences.
- A table when comparing options (providers, auth modes); prose otherwise.
- No internals: no package, type or function names, no file layout beyond the config files a
  user may open themselves.
- Plain second person, present tense, short sentences. State limits plainly ("a secret that
  is still in use cannot be deleted") and say what conm does instead.

## Demos — `docs/demo/`

Each GIF is recorded from the `.tape` file of the same name with
[VHS](https://github.com/charmbracelet/vhs), through `record.sh`:

```sh
docs/demo/record.sh secrets-add.tape
```

`record.sh` starts the databases and the SSH host from `compose.yml`, builds conm from the
working tree, points `XDG_CONFIG_HOME`/`XDG_DATA_HOME` at a temporary directory and puts
`stubs/` first on `PATH`, so a recording never touches the real config. It needs Docker and
`vhs` — `make tools` installs the pinned one into `bin/tools`, which `record.sh` puts on
`PATH` — plus `ttyd` and `ffmpeg`, which VHS drives, and clears its keyring entries through
macOS `security`.

- One tape per behaviour shown, named `<area>-<what>.tape` (`secrets-add`, `transport-use`,
  `workspace-shell`); its `Output` is the same name with `.gif`.
- Copy the `Set` block (shell, font size, 1280×760, theme, typing speed) from an existing
  tape so every GIF looks the same.
- Seed the state the demo starts from inside a `Hide` … `Show` block, so the recording opens
  on the first thing worth watching.
- Embed it under the paragraph it illustrates, with alt text that says what happens:
  `![Adding a secret to the keyring](../demo/secrets-add.gif)`.
- A change to a screen's layout re-records every tape that shows that screen.
