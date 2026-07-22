# Adapters

Per-agent glue. All real logic lives in [`../instructions/CORE.md`](../instructions/CORE.md);
adapters only tell an agent *when* to reach for vellum and *where* the canonical
instructions are. Duplicating logic here is a bug.

Two shapes, because agents load instructions differently:

- **`claude-code/`** — Claude Code uses an **auto-invoked skill**. The adapter is a
  self-contained skill directory (`SKILL.md` + a copy of `CORE.md`,
  `ensure-server.sh`, and the template) installed to `~/.claude/skills/vellum/`.

- **`AGENTS.snippet.md`** — Codex, OpenCode, and Copilot CLI read an **always-on
  `AGENTS.md`-style instruction file**. There is no per-agent content to write: the
  same snippet is inserted (as a managed `<!-- vellum:begin -->…<!-- vellum:end -->`
  block) into each agent's instruction file, and the shared runtime
  (`CORE.md`, `ensure-server.sh`, `templates/base.html`) is installed once to
  `~/.vellum/`, which the snippet points at.

## Install

```sh
./install.sh --agent claude     # skill  -> ~/.claude/skills/vellum/
./install.sh --agent codex      # block  -> ~/.codex/AGENTS.md              (or $CODEX_HOME)
./install.sh --agent opencode   # block  -> ~/.config/opencode/AGENTS.md
./install.sh --agent copilot    # block  -> ~/.copilot/copilot-instructions.md (or $COPILOT_HOME)
```

Re-running is idempotent — the managed block is replaced in place, and any other
instructions in the file are left untouched. In all cases the `vellum` binary is
fetched lazily on first use (or eagerly with `--with-binary`).

The `codex/`, `opencode/`, and `copilot-cli/` directories are intentionally empty:
these agents need no per-agent files beyond the shared snippet above.
