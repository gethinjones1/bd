package picker

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		out  string
		exit int
		want Choice
	}{
		{name: "enter on existing note", out: "kaf\nProjects/Kafka.md\tProjects/Kafka.md\n", exit: 0,
			want: Choice{Kind: Existing, Path: "Projects/Kafka.md"}},
		{name: "enter with empty query picks highlighted", out: "\nDaily/2026-10-10.md\t📅 Today\n", exit: 0,
			want: Choice{Kind: Existing, Path: "Daily/2026-10-10.md"}},
		{name: "no match creates from query", out: "Ideas/Rust CLI\n", exit: 1,
			want: Choice{Kind: New, Path: "Ideas/Rust CLI"}},
		{name: "query is trimmed", out: "  spaced  \n", exit: 1,
			want: Choice{Kind: New, Path: "spaced"}},
		{name: "esc falls back to daily", out: "half typed\n", exit: 130,
			want: Choice{Kind: Daily}},
		{name: "blank query and no selection is daily", out: "\n", exit: 1,
			want: Choice{Kind: Daily}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.out, tt.exit)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseFzfErrorIsAnError(t *testing.T) {
	if _, err := Parse("", 2); err == nil {
		t.Error("want error for fzf exit 2")
	}
}
