package pipeline

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestMathExtension_Inline(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("The area $A = \\pi r^2$ of a circle."), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>The area <span class="math-inline">A = \pi r^2</span> of a circle.</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_Block(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	src := "$$\nA = \\pi r^2\n$$\n"
	if err := md.Convert([]byte(src), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<div class="math-block">A = \pi r^2` + "\n" + `</div>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_EscapedDollar(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("Price is \\$5."), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>Price is $5.</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_NoNewlineInInline(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	src := "The area $A = \\pi\nr^2$ of a circle."
	if err := md.Convert([]byte(src), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	// Should NOT create math-inline because of the newline inside.
	if bytes.Contains(buf.Bytes(), []byte(`class="math-inline"`)) {
		t.Errorf("inline math should not cross newline, got %q", got)
	}
}

func TestMathExtension_CurrencyNotMath(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("Price is $5."), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>Price is $5.</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_CurrencyRangeNotMath(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("Price range is $5 to $10."), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>Price range is $5 to $10.</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_WhitespaceGuard(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("The variable $ a $ is used."), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>The variable $ a $ is used.</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_PureNumberMath(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("Solve $1+2=3$ please."), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>Solve <span class="math-inline">1+2=3</span> please.</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_PureSymbolMath(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("Sign is $+$ ."), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>Sign is <span class="math-inline">+</span> .</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_MultipleInline(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	if err := md.Convert([]byte("text $a$ middle $b$ end"), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<p>text <span class="math-inline">a</span> middle <span class="math-inline">b</span> end</p>` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMathExtension_EmptyBlock(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	var buf bytes.Buffer
	src := "$$\n$$\n"
	if err := md.Convert([]byte(src), &buf); err != nil {
		t.Fatalf("convert error: %v", err)
	}
	got := buf.String()
	want := `<div class="math-block"></div>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
