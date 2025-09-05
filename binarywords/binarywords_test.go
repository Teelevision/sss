package binarywords_test

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/teelevision/sss/binarywords"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		data []byte
		want []string
	}{
		{
			data: nil,
			want: nil,
		},
		{
			data: []byte{},
			want: []string{},
		},
		{
			data: []byte{0x00},
			want: []string{"abend"},
		},
		{
			data: []byte{0x01},
			want: []string{"ablehnen"},
		},
		{
			data: []byte{0xff},
			want: []string{"zwilling"},
		},
		{
			data: []byte{0xff, 0xff},
			want: []string{"zyklus", "ziffer"},
		},
		{
			data: []byte{0b_1111_1111, 0b_1110_0000},
			want: []string{"zyklus", "ader"},
		},
		{
			data: []byte{0x12, 0x34},
			want: []string{"aufstand", "paket"},
		},
		{
			data: []byte{0xde, 0xad, 0xbe, 0xef},
			want: []string{"toilette", "hygiene", "riegel"},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%x", tt.data), func(t *testing.T) {
			if got := binarywords.Encode(tt.data); !slices.Equal(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		words []string
		want  []byte
	}{
		{
			words: nil,
			want:  nil,
		},
		{
			words: []string{},
			want:  []byte{},
		},
		{
			words: []string{"abend"},
			want:  []byte{0x00},
		},
		{
			words: []string{"ablehnen"},
			want:  []byte{0x01},
		},
		{
			words: []string{"zwilling"},
			want:  []byte{0xff},
		},
		{
			words: []string{"zyklus", "ziffer"},
			want:  []byte{0xff, 0xff},
		},
		{
			words: []string{"zyklus", "ader"},
			want:  []byte{0b_1111_1111, 0b_1110_0000},
		},
		{
			words: []string{"aufstand", "paket"},
			want:  []byte{0x12, 0x34},
		},
		{
			words: []string{"toilette", "hygiene", "riegel"},
			want:  []byte{0xde, 0xad, 0xbe, 0xef},
		},
		{
			words: []string{"gibtsnicht"},
			want:  []byte{0xff},
		},
		{
			words: []string{"gibtsnicht", "hygiene", "riegel"},
			want:  []byte{0xff, 0xed, 0xbe, 0xef},
		},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.words, "_"), func(t *testing.T) {
			if got := binarywords.Decode(tt.words); !slices.Equal(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func FuzzEncodeDecode(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		words := binarywords.Encode(data)
		got := binarywords.Decode(words)
		if !slices.Equal(data, got) {
			t.Errorf("roundtrip failed: input=%x, got=%x", data, got)
		}
	})
}
