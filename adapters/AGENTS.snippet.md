## vellum — rich HTML artifacts

When a deliverable is easier to grasp visually than as chat markdown — a
comparison or "X vs Y" trade-off table, a phased plan or roadmap, an
architecture/flow/sequence diagram, a structured report or audit, or a dense
reference table — produce a **vellum artifact** instead of plain markdown.

Follow the canonical instructions in `~/.vellum/CORE.md` exactly: when to
produce an artifact, the output contract (a self-contained HTML file in the
global store `~/.vellum/artifacts/<id>.html`, built from
`~/.vellum/templates/base.html`), the content primitives, optional Mermaid
diagrams, how to open the result via `~/.vellum/ensure-server.sh`, and how to
apply annotations sent back from the viewer.

To view an existing Markdown file (a plan, spec, or doc) as an artifact instead
of authoring a new one, run `vellum render --title "<title>" <path.md>` (flags
before the path) and open the `/view/<id>` URL it prints.

Trigger on the *shape* of the deliverable, not on the words "artifact" or
"HTML". All real logic lives in `~/.vellum/CORE.md`; this block only says when
to reach for it.
