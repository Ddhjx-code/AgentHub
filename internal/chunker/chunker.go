package chunker

import (
	"strings"
	"unicode/utf8"
)

var separators = []string{"\n\n", "\n", "。", ".", "！", "!", "？", "?", "；", ";", " "}

func Split(text string, size, overlap int) []string {
	if size <= 0 {
		size = 512
	}
	if overlap < 0 || overlap >= size {
		overlap = 0
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if utf8.RuneCountInString(text) <= size {
		return []string{text}
	}

	return splitRecursive(text, size, overlap, 0)
}

func splitRecursive(text string, size, overlap, sepIdx int) []string {
	if utf8.RuneCountInString(text) <= size {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return nil
		}
		return []string{trimmed}
	}

	sep := ""
	if sepIdx < len(separators) {
		sep = separators[sepIdx]
	}

	if sep == "" {
		return splitByRune(text, size, overlap)
	}

	parts := strings.Split(text, sep)
	if len(parts) == 1 {
		return splitRecursive(text, size, overlap, sepIdx+1)
	}

	var chunks []string
	current := ""

	for _, part := range parts {
		candidate := current
		if candidate != "" {
			candidate += sep
		}
		candidate += part

		if utf8.RuneCountInString(candidate) > size && current != "" {
			chunks = append(chunks, strings.TrimSpace(current))
			if overlap > 0 {
				current = overlapText(current, overlap) + sep + part
			} else {
				current = part
			}
		} else {
			current = candidate
		}
	}

	if trimmed := strings.TrimSpace(current); trimmed != "" {
		chunks = append(chunks, trimmed)
	}

	var result []string
	for _, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > size {
			result = append(result, splitRecursive(chunk, size, overlap, sepIdx+1)...)
		} else {
			result = append(result, chunk)
		}
	}

	return result
}

func splitByRune(text string, size, overlap int) []string {
	runes := []rune(text)
	var chunks []string
	step := size - overlap
	if step <= 0 {
		step = 1
	}

	for i := 0; i < len(runes); i += step {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[i:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
	}

	return chunks
}

func overlapText(text string, overlap int) string {
	runes := []rune(text)
	if len(runes) <= overlap {
		return text
	}
	return string(runes[len(runes)-overlap:])
}
