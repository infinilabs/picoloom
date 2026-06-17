package pipeline

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// --- AST Nodes ---

// MathInline represents an inline math expression: $...$
type MathInline struct {
	ast.BaseInline
}

// KindMathInline is the node kind for inline math.
var KindMathInline = ast.NewNodeKind("MathInline")

func (n *MathInline) Kind() ast.NodeKind {
	return KindMathInline
}

func (n *MathInline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// MathBlock represents a display math expression: $$...$$
type MathBlock struct {
	ast.BaseBlock
}

// KindMathBlock is the node kind for display math.
var KindMathBlock = ast.NewNodeKind("MathBlock")

func (n *MathBlock) Kind() ast.NodeKind {
	return KindMathBlock
}

func (n *MathBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// --- Inline Parser ($...$) ---

type mathInlineParser struct{}

// NewMathInlineParser creates a parser for inline math delimited by $.
func NewMathInlineParser() parser.InlineParser {
	return &mathInlineParser{}
}

func (p *mathInlineParser) Trigger() []byte {
	return []byte{'$'}
}

func (p *mathInlineParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	if len(line) == 0 {
		return nil
	}

	// We are positioned at the opening '$'.  Search forward for the closing
	// unescaped '$' while enforcing standard math-markdown delimiter guards.
	start := 1
	for i := start; i < len(line); i++ {
		if line[i] == '$' && (i == 0 || line[i-1] != '\\') {
			inner := line[start:i]

			// Guard: no newline inside inline math.
			for _, b := range inner {
				if b == '\n' {
					return nil
				}
			}

			// Guard: opening '$' must not be followed by whitespace.
			if len(inner) > 0 && (inner[0] == ' ' || inner[0] == '\t') {
				return nil
			}
			// Guard: closing '$' must not be preceded by whitespace.
			if len(inner) > 0 {
				last := inner[len(inner)-1]
				if last == ' ' || last == '\t' {
					return nil
				}
			}
			// Guard: opening '$' must not be immediately preceded by a digit
			// (prevents "$5" from being parsed as math).
			prev := block.PrecendingCharacter()
			if prev >= '0' && prev <= '9' {
				return nil
			}
			// Guard: closing '$' must not be immediately followed by a digit.
			nextPos := segment.Start + i + 1
			if nextPos < len(block.Source()) {
				next := block.Source()[nextPos]
				if next >= '0' && next <= '9' {
					return nil
				}
			}

			// Advance reader past the closing $.
			block.Advance(i + 1)
			node := &MathInline{}
			node.AppendChild(node, ast.NewRawTextSegment(text.NewSegment(segment.Start+1, segment.Start+i)))
			return node
		}
	}
	return nil
}

// --- Block Parser ($$...$$) ---

type mathBlockParser struct{}

// NewMathBlockParser creates a parser for display math delimited by $$.
func NewMathBlockParser() parser.BlockParser {
	return &mathBlockParser{}
}

func (b *mathBlockParser) Trigger() []byte {
	return []byte{'$'}
}

func (b *mathBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := util.FirstNonSpacePosition(line)
	if pos < 0 || len(line)-pos < 2 {
		return nil, parser.NoChildren
	}
	if line[pos] != '$' || line[pos+1] != '$' {
		return nil, parser.NoChildren
	}
	// Nothing but whitespace after $$ on the opening line.
	if len(util.TrimRightSpace(line[pos+2:])) > 0 {
		return nil, parser.NoChildren
	}
	node := &MathBlock{}
	node.Lines().Append(segment)
	return node, parser.NoChildren
}

func (b *mathBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	pos := util.FirstNonSpacePosition(line)
	if pos >= 0 && len(line)-pos >= 2 && line[pos] == '$' && line[pos+1] == '$' {
		// End delimiter must have nothing but whitespace after it.
		if len(util.TrimRightSpace(line[pos+2:])) == 0 {
			reader.Advance(segment.Len())
			return parser.Close
		}
	}
	// Accumulate the line into the block.
	node.Lines().Append(segment)
	reader.Advance(segment.Len())
	return parser.Continue | parser.NoChildren
}

func (b *mathBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	lines := node.Lines()
	if lines.Len() > 0 {
		// Drop the opening $$ line (index 0).
		lines.SetSliced(1, lines.Len())
	}
}

func (b *mathBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *mathBlockParser) CanAcceptIndentedLine() bool {
	return false
}

// --- Renderer ---

type mathRenderer struct{}

// NewMathRenderer creates a renderer for math nodes.
func NewMathRenderer() renderer.NodeRenderer {
	return &mathRenderer{}
}

func (r *mathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathInline, r.renderMathInline)
	reg.Register(KindMathBlock, r.renderMathBlock)
}

func (r *mathRenderer) renderMathInline(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<span class="math-inline">`)
	} else {
		_, _ = w.WriteString(`</span>`)
	}
	return ast.WalkContinue, nil
}

func (r *mathRenderer) renderMathBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<div class="math-block">`)
		// MathBlock stores its content in Lines(); emit it directly.
		if block, ok := n.(*MathBlock); ok {
			for i := 0; i < block.Lines().Len(); i++ {
				line := block.Lines().At(i)
				_, _ = w.Write(line.Value(source))
			}
		}
		_, _ = w.WriteString(`</div>`)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

// --- Extension ---

// MathExtension enables $...$ inline and $$...$$ display math support.
type MathExtension struct{}

// NewMathExtension creates the math extension.
func NewMathExtension() goldmark.Extender {
	return &MathExtension{}
}

func (e *MathExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewMathInlineParser(), 500),
		),
		parser.WithBlockParsers(
			util.Prioritized(NewMathBlockParser(), 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewMathRenderer(), 500),
		),
	)
}
