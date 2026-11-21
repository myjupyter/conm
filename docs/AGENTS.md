# `docs/` — documentation

Two audiences, two sets of rules. Read the guide for the kind of documentation you are about
to write before writing it.

| Audience | Where it lives | How to write it |
|---|---|---|
| People using conm | `README.md`, `docs/user/*.md`, the GIFs in `docs/demo/` | `docs/user/AGENTS.md` |
| Agents changing conm | the root `AGENTS.md`, `internal/AGENTS.md`, one `AGENTS.md` per module | `docs/agents/AGENTS.md` |

Which one a change needs:

- **User-visible behaviour changed** — a command, a flag, a screen, a field, a key, a file
  format — the user guide changes in the same piece of work.
- **A module's structure, entities, rules or extension seam changed** — that module's
  `AGENTS.md` changes in the same piece of work.
- Most features need both.

Nothing else belongs in `docs/`: there is no third kind of document. A design note that
explains *why the code is shaped this way* is agent documentation and goes into the module
it is about.
