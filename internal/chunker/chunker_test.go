package chunker

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitEmpty(t *testing.T) {
	result := Split("", 100, 0)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}

	result = Split("   ", 100, 0)
	if result != nil {
		t.Errorf("expected nil for whitespace, got %v", result)
	}
}

func TestSplitShortText(t *testing.T) {
	result := Split("hello world", 100, 0)
	if len(result) != 1 || result[0] != "hello world" {
		t.Errorf("expected single chunk, got %v", result)
	}
}

func TestSplitByParagraph(t *testing.T) {
	text := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
	result := Split(text, 30, 0)

	if len(result) != 3 {
		t.Fatalf("expected 3 chunks, got %d: %v", len(result), result)
	}
	if result[0] != "First paragraph." {
		t.Errorf("chunk 0: %q", result[0])
	}
	if result[1] != "Second paragraph." {
		t.Errorf("chunk 1: %q", result[1])
	}
}

func TestSplitByNewline(t *testing.T) {
	text := "Line one.\nLine two.\nLine three.\nLine four."
	result := Split(text, 25, 0)

	for _, chunk := range result {
		if utf8.RuneCountInString(chunk) > 25 {
			t.Errorf("chunk exceeds size: %q (%d runes)", chunk, utf8.RuneCountInString(chunk))
		}
	}

	joined := strings.Join(result, " ")
	if !strings.Contains(joined, "Line one") || !strings.Contains(joined, "Line four") {
		t.Errorf("missing content in chunks: %v", result)
	}
}

func TestSplitWithOverlap(t *testing.T) {
	text := "AAAA.\n\nBBBB.\n\nCCCC.\n\nDDDD."
	result := Split(text, 10, 5)

	if len(result) < 2 {
		t.Fatalf("expected multiple chunks, got %d: %v", len(result), result)
	}

	for _, chunk := range result {
		if utf8.RuneCountInString(chunk) > 10 {
			t.Errorf("chunk exceeds size: %q", chunk)
		}
	}
}

func TestSplitChinese(t *testing.T) {
	text := "这是第一段内容。这是第二段内容。这是第三段内容。"
	result := Split(text, 10, 0)

	for _, chunk := range result {
		if utf8.RuneCountInString(chunk) > 10 {
			t.Errorf("chunk exceeds size: %q (%d runes)", chunk, utf8.RuneCountInString(chunk))
		}
	}

	if len(result) < 2 {
		t.Errorf("expected multiple chunks for Chinese text, got %d", len(result))
	}
}

func TestSplitLongLine(t *testing.T) {
	text := strings.Repeat("a", 100)
	result := Split(text, 30, 0)

	if len(result) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(result))
	}

	for _, chunk := range result {
		if utf8.RuneCountInString(chunk) > 30 {
			t.Errorf("chunk exceeds size: %d runes", utf8.RuneCountInString(chunk))
		}
	}
}

func TestSplitDefaultSize(t *testing.T) {
	text := strings.Repeat("word ", 200)
	result := Split(text, 0, 0)

	if len(result) < 1 {
		t.Fatal("expected at least 1 chunk")
	}

	for _, chunk := range result {
		if utf8.RuneCountInString(chunk) > 512 {
			t.Errorf("chunk exceeds default size: %d runes", utf8.RuneCountInString(chunk))
		}
	}
}

func TestSplitInvalidOverlap(t *testing.T) {
	text := "hello world foo bar baz"
	result := Split(text, 10, 10)
	if len(result) < 1 {
		t.Fatal("expected at least 1 chunk with overlap >= size")
	}
}
