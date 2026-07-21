// Package markdown is a tiny, dependency-free Markdown-to-HTML renderer for the
// constructs used in project docs: headings, fenced code, inline code, GFM
// tables, ordered/unordered lists (incl. task checkboxes), blockquotes,
// horizontal rules, bold, and links. Output is plain semantic HTML styled by
// base.html's element-level CSS. All text is HTML-escaped.
package markdown

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	reHeading  = regexp.MustCompile(`^(#{1,4})\s+(.*)$`)
	reHR       = regexp.MustCompile(`^---+\s*$`)
	reList     = regexp.MustCompile(`^\s*([-*]|\d+\.)\s+`)
	reOrdered  = regexp.MustCompile(`^\s*\d+\.\s+`)
	reTask     = regexp.MustCompile(`^\[([ xX])\]\s+(.*)$`)
	reInlCode  = regexp.MustCompile("`([^`]+)`")
	reBold     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reLink     = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reTableSep = regexp.MustCompile(`^\s*\|?[\s:|-]+\|?\s*$`)
)

func inlineFmt(s string) string {
	s = html.EscapeString(s)
	s = reInlCode.ReplaceAllString(s, "<code>$1</code>")
	s = reBold.ReplaceAllString(s, "<strong>$1</strong>")
	s = reLink.ReplaceAllString(s, `<a href="$2">$1</a>`)
	return s
}

func cells(row string) []string {
	row = strings.Trim(strings.TrimSpace(row), "|")
	parts := strings.Split(row, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// Render converts Markdown to an HTML fragment.
func Render(md string) string {
	lines := strings.Split(md, "\n")
	var out, para []string
	flush := func() {
		if len(para) > 0 {
			out = append(out, "<p>"+inlineFmt(strings.Join(para, " "))+"</p>")
			para = nil
		}
	}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		switch {
		case strings.HasPrefix(line, "```"):
			flush()
			i++
			var code []string
			for i < len(lines) && !strings.HasPrefix(lines[i], "```") {
				code = append(code, lines[i])
				i++
			}
			out = append(out, "<pre><code>"+html.EscapeString(strings.Join(code, "\n"))+"</code></pre>")
		case reHeading.MatchString(line):
			flush()
			m := reHeading.FindStringSubmatch(line)
			lvl := len(m[1])
			out = append(out, fmt.Sprintf("<h%d>%s</h%d>", lvl, inlineFmt(m[2]), lvl))
		case reHR.MatchString(line):
			flush()
			out = append(out, "<hr>")
		case strings.HasPrefix(strings.TrimSpace(line), "|") && i+1 < len(lines) && reTableSep.MatchString(lines[i+1]):
			flush()
			header := cells(line)
			i += 2
			var body [][]string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
				body = append(body, cells(lines[i]))
				i++
			}
			i--
			var b strings.Builder
			b.WriteString(`<div class="table-scroll"><table><thead><tr>`)
			for _, c := range header {
				b.WriteString("<th>" + inlineFmt(c) + "</th>")
			}
			b.WriteString("</tr></thead><tbody>")
			for _, row := range body {
				b.WriteString("<tr>")
				for _, c := range row {
					b.WriteString("<td>" + inlineFmt(c) + "</td>")
				}
				b.WriteString("</tr>")
			}
			b.WriteString("</tbody></table></div>")
			out = append(out, b.String())
		case strings.HasPrefix(line, ">"):
			flush()
			var q []string
			for i < len(lines) && strings.HasPrefix(lines[i], ">") {
				q = append(q, strings.TrimSpace(strings.TrimPrefix(lines[i], ">")))
				i++
			}
			i--
			out = append(out, "<blockquote>"+inlineFmt(strings.Join(q, " "))+"</blockquote>")
		case reList.MatchString(line):
			flush()
			ordered := reOrdered.MatchString(line)
			var items []string
			for i < len(lines) && reList.MatchString(lines[i]) {
				item := reList.ReplaceAllString(lines[i], "")
				if cb := reTask.FindStringSubmatch(item); cb != nil {
					mark := "☐"
					if strings.EqualFold(cb[1], "x") {
						mark = "☑"
					}
					items = append(items, `<li><span class="badge badge--muted">`+mark+"</span> "+inlineFmt(cb[2])+"</li>")
				} else {
					items = append(items, "<li>"+inlineFmt(item)+"</li>")
				}
				i++
			}
			i--
			tag := "ul"
			if ordered {
				tag = "ol"
			}
			out = append(out, "<"+tag+">"+strings.Join(items, "")+"</"+tag+">")
		case strings.TrimSpace(line) == "":
			flush()
		default:
			para = append(para, strings.TrimSpace(line))
		}
	}
	flush()
	return strings.Join(out, "\n")
}
