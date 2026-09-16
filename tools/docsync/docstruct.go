package main

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Section is one heading-bounded block of a documentation page, in document
// order. Nested headings become child sections; paragraphs, list items and
// code blocks accumulate in Texts; the first table of the section (if any) is
// kept in Table.
type Section struct {
	Level    int
	Heading  string
	Texts    []string
	Table    *Table
	Children []*Section
}

// Table is a documentation table with headers in column order.
type Table struct {
	Headers []string
	Rows    [][]string
}

// endpointRe matches "GET https://api.weixin.qq.com/xxx" style declarations
// in the page text.
var endpointRe = regexp.MustCompile(`\b(GET|POST|PUT|DELETE|PATCH)\s+((?:https?://api\.weixin\.qq\.com)?/[A-Za-z0-9_\-./?=&%{}]+)`)

// endpointNoisyRe is the fallback for pages whose call URL carries a stray
// word between the verb and the address ("POST New https://api.weixin.qq.com/
// tcb/invokecloudfunction"). Exactly one short ASCII word may be interposed,
// so ordinary prose does not start matching.
var endpointNoisyRe = regexp.MustCompile(`\b(GET|POST|PUT|DELETE|PATCH)\s+[A-Za-z]{1,6}\s+((?:https?://api\.weixin\.qq\.com)?/[A-Za-z0-9_\-./?=&%{}]+)`)

// findEndpointMatches returns the endpoint declarations in text, falling back
// to the noisy form when the strict pattern finds nothing.
func findEndpointMatches(text string) [][]string {
	if m := endpointRe.FindAllStringSubmatch(text, -1); len(m) > 0 {
		return m
	}
	return endpointNoisyRe.FindAllStringSubmatch(text, -1)
}

// trailingPunct strips punctuation that often leaks from inline code text.
var trailingPunct = regexp.MustCompile(`[.,;:)]+$`)

// parseStructured converts the content block of a doc page into a heading-
// nested section tree. The returned root is a synthetic level-0 section whose
// Children are the top-level sections in document order.
func parseStructured(content *html.Node) *Section {
	root := &Section{Level: 0}
	cur := root
	stack := []*Section{root}

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type != html.ElementNode {
			if n.Type == html.TextNode {
				appendText(&cur.Texts, n.Data)
			}
			return
		}
		switch strings.ToLower(n.Data) {
		case "script", "style", "template", "noscript":
			return
		case "h1", "h2", "h3", "h4", "h5", "h6":
			level := int(n.Data[1] - '0')
			// headings carry a decorative "# " header-anchor; drop it
			heading := strings.TrimLeft(headingText(n), "# \t")
			sec := &Section{Level: level, Heading: heading}
			for len(stack) > 1 && stack[len(stack)-1].Level >= level {
				stack = stack[:len(stack)-1]
			}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, sec)
			stack = append(stack, sec)
			cur = sec
		case "ul", "ol":
			for li := n.FirstChild; li != nil; li = li.NextSibling {
				if li.Type == html.ElementNode && strings.EqualFold(li.Data, "li") {
					appendText(&cur.Texts, headingText(li))
				}
			}
		case "table":
			if t := parseTable(n); t != nil {
				if cur.Table == nil {
					cur.Table = t
				} else {
					cur.Texts = append(cur.Texts, renderTableText(t))
				}
			}
		case "pre":
			if txt := preText(n); txt != "" {
				cur.Texts = append(cur.Texts, txt)
			}
		case "br":
			// pure layout; nothing to record
		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	for c := content.FirstChild; c != nil; c = c.NextSibling {
		walk(c)
	}
	return root
}

// appendText appends a whitespace-collapsed, deduplicated text entry.
func appendText(texts *[]string, raw string) {
	txt := strings.Join(strings.Fields(raw), " ")
	if txt == "" {
		return
	}
	if len(*texts) > 0 && (*texts)[len(*texts)-1] == txt {
		return
	}
	*texts = append(*texts, txt)
}

// parseTable extracts headers (in column order) and cell rows from a table
// element. Tables without header cells get nil Headers (emitted as arrays).
func parseTable(n *html.Node) *Table {
	var headers []string
	var rows [][]string

	var handleRow func(tr *html.Node, isHeader bool) bool
	handleRow = func(tr *html.Node, isHeader bool) bool {
		var cells []string
		for cell := tr.FirstChild; cell != nil; cell = cell.NextSibling {
			if cell.Type != html.ElementNode {
				continue
			}
			if !strings.EqualFold(cell.Data, "td") && !strings.EqualFold(cell.Data, "th") {
				continue
			}
			cells = append(cells, headingText(cell))
		}
		if len(cells) == 0 {
			return false
		}
		if isHeader {
			headers = append(headers, cells...)
		} else {
			rows = append(rows, cells)
		}
		return true
	}

	var walkRows func(n *html.Node, inHead bool) bool
	walkRows = func(n *html.Node, inHead bool) bool {
		foundHead := false
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode {
				continue
			}
			switch strings.ToLower(c.Data) {
			case "thead":
				for tr := c.FirstChild; tr != nil; tr = tr.NextSibling {
					if tr.Type == html.ElementNode && strings.EqualFold(tr.Data, "tr") {
						if handleRow(tr, true) {
							foundHead = true
						}
					}
				}
			case "tbody", "tfoot":
				foundHead = walkRows(c, inHead) || foundHead
			case "tr":
				if handleRow(c, inHead && !foundHead) {
					foundHead = inHead || foundHead
				}
			default:
				foundHead = walkRows(c, inHead) || foundHead
			}
		}
		return foundHead
	}
	walkRows(n, false)

	// Some tables mark up their header row with plain td cells. When the
	// first row consists entirely of well-known header vocabulary, promote
	// it so rows still become ordered-key objects.
	if headers == nil && len(rows) > 0 && looksLikeHeaderRow(rows[0]) {
		headers = rows[0]
		rows = rows[1:]
	}
	if headers == nil && rows == nil {
		return nil
	}
	return &Table{Headers: headers, Rows: rows}
}

// headerVocabulary covers the header cells used by the WeChat doc tables.
var headerVocabulary = map[string]bool{
	"参数名": true, "参数": true, "属性": true, "名称": true, "字段": true,
	"类型": true, "必填": true, "描述": true, "说明": true, "含义": true,
	"默认值": true, "默认": true, "示例值": true, "示例": true, "返回值": true,
	"返回类型": true, "取值": true, "错误码": true, "错误描述": true,
	"错误信息": true, "解决方案": true, "排查方法": true, "errmsg": true,
	// optional columns the docs sometimes add
	"枚举": true, "枚举值": true, "取值范围": true, "单位": true,
	"最小值": true, "最大值": true, "备注": true, "是否必填": true,
	"name": true, "type": true, "required": true, "default": true,
	"description": true, "enum": true, "example": true,
}

func looksLikeHeaderRow(cells []string) bool {
	if len(cells) < 2 {
		return false
	}
	for _, cell := range cells {
		if !headerVocabulary[strings.ToLower(strings.TrimSpace(cell))] {
			return false
		}
	}
	return true
}

// renderTableText flattens an extra table of a section into a text entry so
// its content is not lost.
func renderTableText(t *Table) string {
	var parts []string
	for _, row := range t.Rows {
		parts = append(parts, strings.Join(row, " | "))
	}
	return strings.Join(parts, "\n")
}

// preText returns the text of a <pre> block with original line structure but
// trimmed edges.
func preText(n *html.Node) string {
	var sb strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	lines := strings.Split(sb.String(), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}
	return strings.Trim(strings.Join(lines, "\n"), "\n")
}
