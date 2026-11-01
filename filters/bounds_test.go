package filters

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestFilterBounds(t *testing.T) {
	cases := []struct {
		name    string
		call    func() (interface{}, error)
		wantErr string // "" means err must be nil; "*" means any error
		wantOut string // checked via %v when wantErr == ""
	}{
		{"truncate_negative", func() (interface{}, error) { return TruncateFilter("hello world", -1) }, "*", ""},
		{"truncatehtml_negative", func() (interface{}, error) { return TruncateHTMLFilter("hello world", -1) }, "*", ""},
		{"center_huge", func() (interface{}, error) { return CenterFilter("x", 1<<62) }, "too large", ""},
		{"center_over_cap", func() (interface{}, error) { return CenterFilter("x", 100<<20) }, "too large", ""},
		{"pad_left_huge", func() (interface{}, error) { return PadLeftFilter("x", 1<<62) }, "too large", ""},
		{"pad_right_huge", func() (interface{}, error) { return PadRightFilter("x", 1<<62) }, "too large", ""},
		{"indent_huge", func() (interface{}, error) { return IndentFilter("a\nb", 1<<62) }, "too large", ""},
		{"indent_amplified", func() (interface{}, error) { return IndentFilter(strings.Repeat("x\n", 1000), 100000) }, "too large", ""},
		{"indent_negative", func() (interface{}, error) { return IndentFilter("a\nb", -1) }, "", "a\nb"},
		{"batch_huge_no_fill", func() (interface{}, error) { return BatchFilter([]interface{}{1, 2, 3}, 1<<62) }, "", "[[1 2 3]]"},
		{"batch_huge_fill", func() (interface{}, error) { return BatchFilter([]interface{}{1}, 1<<62, "x") }, "too large", ""},
		{"batch_fill_over_count", func() (interface{}, error) { return BatchFilter([]interface{}{1}, 2000000, 0) }, "too large", ""},
		{"batch_fill_ok", func() (interface{}, error) { return BatchFilter([]interface{}{1, 2, 3}, 2, 0) }, "", "[[1 2] [3 0]]"},
		{"currency_nan", func() (interface{}, error) { return CurrencyFilter(math.NaN()) }, "finite", ""},
		{"currency_inf", func() (interface{}, error) { return CurrencyFilter(math.Inf(1)) }, "finite", ""},
		{"center_ok", func() (interface{}, error) { return CenterFilter("x", 3) }, "", " x "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out interface{}
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panicked: %v", r)
					}
				}()
				out, err = tc.call()
			}()
			switch tc.wantErr {
			case "":
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got := fmt.Sprintf("%v", out); got != tc.wantOut {
					t.Fatalf("got %q, want %q", got, tc.wantOut)
				}
			case "*":
				if err == nil {
					t.Fatalf("expected error, got %v", out)
				}
			default:
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
			}
		})
	}
}
