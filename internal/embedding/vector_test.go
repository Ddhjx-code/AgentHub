package embedding

import (
	"math"
	"testing"
)

func TestEncodeDecodeFloat32s(t *testing.T) {
	original := []float32{1.0, -2.5, 3.14, 0.0, -0.001}
	encoded := EncodeFloat32s(original)

	if len(encoded) != len(original)*4 {
		t.Fatalf("expected %d bytes, got %d", len(original)*4, len(encoded))
	}

	decoded := DecodeFloat32s(encoded)
	if len(decoded) != len(original) {
		t.Fatalf("expected %d floats, got %d", len(original), len(decoded))
	}

	for i := range original {
		if decoded[i] != original[i] {
			t.Errorf("index %d: expected %f, got %f", i, original[i], decoded[i])
		}
	}
}

func TestEncodeDecodeEmpty(t *testing.T) {
	encoded := EncodeFloat32s(nil)
	if len(encoded) != 0 {
		t.Errorf("expected empty, got %d bytes", len(encoded))
	}

	decoded := DecodeFloat32s(nil)
	if len(decoded) != 0 {
		t.Errorf("expected empty, got %d floats", len(decoded))
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float32
		want float32
	}{
		{
			name: "identical vectors",
			a:    []float32{1, 0, 0},
			b:    []float32{1, 0, 0},
			want: 1.0,
		},
		{
			name: "orthogonal vectors",
			a:    []float32{1, 0, 0},
			b:    []float32{0, 1, 0},
			want: 0.0,
		},
		{
			name: "opposite vectors",
			a:    []float32{1, 0, 0},
			b:    []float32{-1, 0, 0},
			want: -1.0,
		},
		{
			name: "similar vectors",
			a:    []float32{1, 2, 3},
			b:    []float32{1, 2, 3.1},
			want: 0.9999,
		},
		{
			name: "different lengths",
			a:    []float32{1, 2},
			b:    []float32{1, 2, 3},
			want: 0.0,
		},
		{
			name: "empty vectors",
			a:    []float32{},
			b:    []float32{},
			want: 0.0,
		},
		{
			name: "zero vector",
			a:    []float32{0, 0, 0},
			b:    []float32{1, 2, 3},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			if math.Abs(float64(got-tt.want)) > 0.001 {
				t.Errorf("CosineSimilarity(%v, %v) = %f, want %f", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
