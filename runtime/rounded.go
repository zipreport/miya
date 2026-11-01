package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
)

// RoundedFloat is the result of round(precision) for precision >= 1: a number
// that prints with Places decimals (miya's deliberate round deviation).
type RoundedFloat struct {
	Value  float64
	Places int
}

func (r RoundedFloat) String() string { return strconv.FormatFloat(r.Value, 'f', r.Places, 64) }

func (r RoundedFloat) MarshalJSON() ([]byte, error) { return json.Marshal(r.Value) }

// Format makes fmt print the rounded text for %v/%s, and the number otherwise.
func (r RoundedFloat) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		_, _ = io.WriteString(f, r.String())
	case 'd':
		fmt.Fprintf(f, fmt.FormatString(f, verb), int64(math.Trunc(r.Value)))
	default:
		fmt.Fprintf(f, fmt.FormatString(f, verb), r.Value)
	}
}
