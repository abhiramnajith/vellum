package markdown

import (
	"strings"
	"testing"
)

func TestRenderBasics(t *testing.T) {
	md := "# Title\n\nSome **bold** and `code`.\n\n- [ ] todo\n- [x] done\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n```go\nfmt.Println(\"hi\")\n```\n"
	got := Render(md)
	for _, want := range []string{
		"<h1>Title</h1>",
		"<strong>bold</strong>",
		"<code>code</code>",
		"<li>", "todo", "done",
		"<div class=\"table-scroll\"><table>",
		"<pre><code>", "fmt.Println",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Render output missing %q\n---\n%s", want, got)
		}
	}
	if strings.Contains(got, "<script") {
		t.Fatal("Render must not emit raw <script>")
	}
}

func TestRenderEscapesHTML(t *testing.T) {
	got := Render("a <img src=x onerror=alert(1)> b")
	if strings.Contains(got, "<img") {
		t.Fatalf("HTML in markdown text must be escaped, got: %s", got)
	}
}
