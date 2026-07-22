#!/usr/bin/env sh
# install.sh — install the vellum adapter for a coding agent.
#
# Claude Code uses an auto-invoked skill, so its adapter is a self-contained
# skill directory under ~/.claude/skills/vellum/. The other agents (Codex,
# OpenCode, Copilot CLI) read an always-on AGENTS.md-style instruction file, so
# their adapter installs the shared runtime under ~/.vellum/ and inserts a
# marked, idempotent block into that agent's instruction file — without
# clobbering any instructions the user already has there.
#
# Deliberately readable and dumb: no `curl | bash`, no network calls beyond the
# `git clone` you already ran to get here. Re-running it is idempotent.
#
# Usage:
#   ./install.sh --agent claude|codex|opencode|copilot [--local] [--with-binary]
#
#   --agent        which agent to install for
#   --local        Claude only: install into ./.claude/skills/ instead of ~/
#   --with-binary  eagerly fetch the server binary now, instead of on first use

set -eu

AGENT=""
LOCAL=0
WITH_BINARY=0

usage() {
	echo "usage: ./install.sh --agent claude|codex|opencode|copilot [--local] [--with-binary]" >&2
}

while [ $# -gt 0 ]; do
	case "$1" in
	--agent)
		shift
		[ $# -gt 0 ] || { echo "error: --agent needs a value" >&2; usage; exit 2; }
		AGENT="$1"
		;;
	--agent=*)
		AGENT="${1#--agent=}"
		;;
	--local)
		LOCAL=1
		;;
	--with-binary)
		WITH_BINARY=1
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "error: unknown argument: $1" >&2
		usage
		exit 2
		;;
	esac
	shift
done

if [ -z "$AGENT" ]; then
	echo "error: --agent is required" >&2
	usage
	exit 2
fi

# Resolve the repo root (directory containing this script) so the script works
# regardless of the caller's working directory.
REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
CORE="$REPO_ROOT/instructions/CORE.md"
ENSURE="$REPO_ROOT/scripts/ensure-server.sh"
BASE_HTML="$REPO_ROOT/instructions/templates/base.html"
SNIPPET="$REPO_ROOT/adapters/AGENTS.snippet.md"

VELLUM_HOME="${VELLUM_HOME:-$HOME/.vellum}"

# fetch_binary optionally pre-fetches the server binary via ensure-server.sh.
fetch_binary() {
	if [ "$WITH_BINARY" -eq 1 ]; then
		echo "fetching server binary now (--with-binary)..."
		if VELLUM_HOME="$VELLUM_HOME" "$ENSURE" >/dev/null; then
			echo "server binary ready."
		else
			echo "warning: failed to fetch the server binary; it will be fetched on first use instead." >&2
		fi
	else
		echo "note: the server binary is fetched automatically on first use (run with --with-binary to fetch it now)."
	fi
}

# install_runtime copies the shared runtime (CORE.md, ensure-server.sh, the
# base.html template) into $VELLUM_HOME so the AGENTS.md snippet can point at it.
install_runtime() {
	mkdir -p "$VELLUM_HOME/templates"
	cp "$CORE" "$VELLUM_HOME/CORE.md"
	cp "$ENSURE" "$VELLUM_HOME/ensure-server.sh"
	chmod +x "$VELLUM_HOME/ensure-server.sh"
	cp "$BASE_HTML" "$VELLUM_HOME/templates/base.html"
}

# upsert_block inserts (or replaces in place) the vellum instruction block in an
# agent instruction file, leaving any other content untouched. Idempotent.
#   $1 = target instruction file
upsert_block() {
	target="$1"
	begin="<!-- vellum:begin (managed by install.sh — edits between these markers are overwritten) -->"
	end="<!-- vellum:end -->"
	mkdir -p "$(dirname "$target")"
	[ -f "$target" ] || : > "$target"

	tmp="$target.vellum.tmp"
	# Copy everything except a previously-installed block (inclusive of markers).
	awk -v b="$begin" -v e="$end" '
		$0 == b { skip = 1 }
		skip && $0 == e { skip = 0; next }
		!skip { print }
	' "$target" > "$tmp"

	# Trim a trailing blank line so re-runs do not accumulate whitespace.
	# Then append a separating blank line and the fresh block.
	printf '%s\n\n' "$begin" >> "$tmp"
	cat "$SNIPPET" >> "$tmp"
	printf '%s\n' "$end" >> "$tmp"

	mv "$tmp" "$target"
	echo "updated vellum instructions in $target"
}

case "$AGENT" in
claude)
	# Auto-invoked skill: self-contained directory install.
	ADAPTER_DIR="$REPO_ROOT/adapters/claude-code"
	if [ ! -f "$ADAPTER_DIR/SKILL.md" ]; then
		echo "error: no SKILL.md found in $ADAPTER_DIR" >&2
		exit 1
	fi
	if [ "$LOCAL" -eq 1 ]; then
		TARGET="./.claude/skills/vellum"
	else
		TARGET="$HOME/.claude/skills/vellum"
	fi
	mkdir -p "$TARGET/templates"
	cp "$ADAPTER_DIR/SKILL.md" "$TARGET/SKILL.md"
	cp "$CORE" "$TARGET/CORE.md"
	cp "$ENSURE" "$TARGET/ensure-server.sh"
	chmod +x "$TARGET/ensure-server.sh"
	cp "$BASE_HTML" "$TARGET/templates/base.html"
	echo "installed vellum skill for 'claude' -> $TARGET"
	fetch_binary
	;;
codex)
	[ "$LOCAL" -eq 0 ] || echo "note: --local has no effect for codex (global AGENTS.md install)" >&2
	install_runtime
	upsert_block "${CODEX_HOME:-$HOME/.codex}/AGENTS.md"
	fetch_binary
	;;
opencode)
	[ "$LOCAL" -eq 0 ] || echo "note: --local has no effect for opencode (global AGENTS.md install)" >&2
	install_runtime
	upsert_block "$HOME/.config/opencode/AGENTS.md"
	fetch_binary
	;;
copilot | copilot-cli)
	[ "$LOCAL" -eq 0 ] || echo "note: --local has no effect for copilot (global instructions install)" >&2
	install_runtime
	upsert_block "${COPILOT_HOME:-$HOME/.copilot}/copilot-instructions.md"
	fetch_binary
	;;
*)
	echo "error: unknown agent '$AGENT' (expected: claude|codex|opencode|copilot)" >&2
	exit 2
	;;
esac
