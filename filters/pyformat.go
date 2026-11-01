package filters

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/zipreport/miya/runtime"
)

// maxFormatWidth bounds widths and precisions so a template cannot request
// gigabytes of padding.
const maxFormatWidth = 1 << 20

type pyFlags struct {
	alt, zero, left, plus, space bool
}

// pyFormat implements Python's `str % args` (or `str % mapping` when kwargs is
// not empty) for the conversions d i u o x X e E f F g G c s r a and %%.
// With safe set the format string is trusted markup, as in markupsafe: the
// text of %s, %r and %a arguments is HTML-escaped first (safe arguments are
// kept) and %c is an error.
func pyFormat(format string, args []interface{}, kwargs map[string]interface{}, safe bool) (string, error) {
	rs := []rune(format)
	mapping := len(kwargs) > 0
	ai := 0
	mapUsed := false
	var out strings.Builder

	next := func() (interface{}, error) {
		if mapping {
			if mapUsed {
				return nil, pyFormatErr("not enough arguments for format string")
			}
			mapUsed = true
			return map[string]interface{}(kwargs), nil
		}
		if ai >= len(args) {
			return nil, pyFormatErr("not enough arguments for format string")
		}
		ai++
		return args[ai-1], nil
	}
	star := func() (int64, error) {
		a, err := next()
		if err != nil {
			return 0, err
		}
		if mapping {
			return 0, pyFormatErr("* wants int")
		}
		n, err := pyIntArg(runtime.Unmark(a), '*')
		if err != nil {
			return 0, pyFormatErr("* wants int")
		}
		if !n.IsInt64() || n.Int64() > maxFormatWidth || n.Int64() < -maxFormatWidth {
			return 0, &runtime.SizeError{Msg: "format: width too big"}
		}
		return n.Int64(), nil
	}

	for i := 0; i < len(rs); i++ {
		if rs[i] != '%' {
			out.WriteRune(rs[i])
			continue
		}
		i++
		if i >= len(rs) {
			return "", pyFormatErr("incomplete format")
		}
		if rs[i] == '%' {
			out.WriteByte('%')
			continue
		}

		var arg interface{}
		hasKey := false
		if rs[i] == '(' {
			depth := 1
			j := i + 1
			for j < len(rs) && depth > 0 {
				switch rs[j] {
				case '(':
					depth++
				case ')':
					depth--
				}
				j++
			}
			if depth > 0 {
				return "", pyFormatErr("incomplete format key")
			}
			if !mapping {
				return "", pyFormatErr("format requires a mapping")
			}
			key := string(rs[i+1 : j-1])
			v, ok := kwargs[key]
			if !ok {
				return "", pyFormatErr("'" + key + "'")
			}
			arg, hasKey = v, true
			mapUsed = true
			i = j
		}

		var fl pyFlags
	flags:
		for ; i < len(rs); i++ {
			switch rs[i] {
			case '#':
				fl.alt = true
			case '0':
				fl.zero = true
			case '-':
				fl.left = true
			case '+':
				fl.plus = true
			case ' ':
				fl.space = true
			default:
				break flags
			}
		}

		width, prec := 0, -1
		if i < len(rs) && rs[i] == '*' {
			w, err := star()
			if err != nil {
				return "", err
			}
			if w < 0 {
				fl.left = true
				w = -w
			}
			width = int(w)
			i++
		} else {
			for i < len(rs) && rs[i] >= '0' && rs[i] <= '9' {
				width = width*10 + int(rs[i]-'0')
				if width > maxFormatWidth {
					return "", &runtime.SizeError{Msg: "format: width too big"}
				}
				i++
			}
		}
		if i < len(rs) && rs[i] == '.' {
			i++
			prec = 0
			if i < len(rs) && rs[i] == '*' {
				p, err := star()
				if err != nil {
					return "", err
				}
				if p < 0 {
					p = 0
				}
				prec = int(p)
				i++
			} else {
				for i < len(rs) && rs[i] >= '0' && rs[i] <= '9' {
					prec = prec*10 + int(rs[i]-'0')
					if prec > maxFormatWidth {
						return "", &runtime.SizeError{Msg: "format: precision too big"}
					}
					i++
				}
			}
		}
		for i < len(rs) && (rs[i] == 'h' || rs[i] == 'l' || rs[i] == 'L') {
			i++
		}
		if i >= len(rs) {
			return "", pyFormatErr("incomplete format")
		}

		c := rs[i]
		if !strings.ContainsRune("diuoxXeEfFgGcsra", c) {
			return "", pyFormatErr(fmt.Sprintf("unsupported format character %s (0x%x) at index %d", pyQuoteRune(c), c, i))
		}
		if !hasKey {
			a, err := next()
			if err != nil {
				return "", err
			}
			arg = a
		}

		var text string
		switch c {
		case 'd', 'i', 'u', 'o', 'x', 'X':
			n, err := pyIntArg(runtime.Unmark(arg), c)
			if err != nil {
				return "", err
			}
			text = pyFormatInt(n, byte(c), fl, width, prec)
		case 'e', 'E', 'f', 'F', 'g', 'G':
			f, err := pyFloatArg(runtime.Unmark(arg), c)
			if err != nil {
				return "", err
			}
			text = pyFormatFloat(f, byte(c), fl, width, prec)
		case 'c':
			if safe {
				return "", pyFormatErr("%c requires an int or a unicode character, not an escaped argument")
			}
			s, err := pyCharArg(arg)
			if err != nil {
				return "", err
			}
			text = pyPad(s, width, fl.left)
		default:
			var s string
			switch c {
			case 's':
				s = ToString(arg)
				if _, isSafe := runtime.AsMarkup(arg); safe && !isSafe {
					s = string(runtime.EscapeMarkup(s))
				}
			case 'r', 'a':
				if str, ok := runtime.BasicValue(runtime.Unmark(arg)).(string); ok {
					s = pyRepr(str)
					if c == 'a' {
						s = pyASCII(s)
					}
				} else {
					s = ToString(arg)
				}
				if safe {
					s = string(runtime.EscapeMarkup(s))
				}
			}
			if prec >= 0 && utf8.RuneCountInString(s) > prec {
				s = string([]rune(s)[:prec])
			}
			text = pyPad(s, width, fl.left)
		}
		if out.Len()+len(text) > maxRepeatBytes {
			return "", &runtime.SizeError{Msg: "format output too large"}
		}
		out.WriteString(text)
	}

	if !mapping && ai < len(args) {
		return "", pyFormatErr("not all arguments converted during string formatting")
	}
	return out.String(), nil
}

func pyFormatErr(msg string) error { return fmt.Errorf("format: %s", msg) }

func pyQuoteRune(c rune) string {
	if c == '\'' {
		return `"'"`
	}
	return "'" + string(c) + "'"
}

// pyFormatTypeName returns the Python type name of a template value, for messages.
func pyFormatTypeName(v interface{}) string {
	switch v.(type) {
	case nil:
		return "NoneType"
	case string:
		return "str"
	case []interface{}:
		return "list"
	case map[string]interface{}:
		return "dict"
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.String:
		return "str"
	case reflect.Slice, reflect.Array:
		return "list"
	case reflect.Map:
		return "dict"
	}
	return fmt.Sprintf("%T", v)
}

// pyNumber classifies a numeric template value as an integer or a float.
func pyNumber(v interface{}) (n *big.Int, f float64, isFloat, ok bool) {
	v = runtime.Indirect(v)
	switch x := v.(type) {
	case nil:
		return nil, 0, false, false
	case runtime.RoundedFloat:
		return nil, x.Value, true, true
	case bool:
		if x {
			return big.NewInt(1), 0, false, true
		}
		return big.NewInt(0), 0, false, true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return big.NewInt(1), 0, false, true
		}
		return big.NewInt(0), 0, false, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return big.NewInt(rv.Int()), 0, false, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return new(big.Int).SetUint64(rv.Uint()), 0, false, true
	case reflect.Float32, reflect.Float64:
		return nil, rv.Float(), true, true
	}
	return nil, 0, false, false
}

// pyWholeFloat converts a finite float without a fraction to an integer.
func pyWholeFloat(f float64) *big.Int {
	if math.Abs(f) < 1<<63 {
		return big.NewInt(int64(f))
	}
	n, _ := big.NewFloat(f).Int(nil)
	return n
}

// pyIntArg converts an argument for an integer conversion. d, i and u accept
// any number and truncate floats; the others accept integers and whole floats.
func pyIntArg(arg interface{}, c rune) (*big.Int, error) {
	n, f, isFloat, ok := pyNumber(arg)
	if !ok {
		if c == 'd' || c == 'i' || c == 'u' {
			return nil, pyFormatErr(fmt.Sprintf("%%%c format: a real number is required, not %s", c, pyFormatTypeName(arg)))
		}
		return nil, pyFormatErr(fmt.Sprintf("%%%c format: an integer is required, not %s", c, pyFormatTypeName(arg)))
	}
	if !isFloat {
		return n, nil
	}
	if c == 'd' || c == 'i' || c == 'u' {
		if math.IsNaN(f) {
			return nil, pyFormatErr("cannot convert float NaN to integer")
		}
		if math.IsInf(f, 0) {
			return nil, pyFormatErr("cannot convert float infinity to integer")
		}
		return pyWholeFloat(math.Trunc(f)), nil
	}
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return nil, pyFormatErr(fmt.Sprintf("%%%c format: an integer is required, not float", c))
	}
	return pyWholeFloat(f), nil
}

func pyFloatArg(arg interface{}, c rune) (float64, error) {
	n, f, isFloat, ok := pyNumber(arg)
	if !ok {
		return 0, pyFormatErr(fmt.Sprintf("must be real number, not %s", pyFormatTypeName(arg)))
	}
	if isFloat {
		return f, nil
	}
	f, _ = new(big.Float).SetInt(n).Float64()
	return f, nil
}

func pyCharArg(arg interface{}) (string, error) {
	if s, ok := arg.(string); ok {
		if utf8.RuneCountInString(s) != 1 {
			return "", pyFormatErr(fmt.Sprintf("%%c requires an int or a unicode character, not a string of length %d", utf8.RuneCountInString(s)))
		}
		return s, nil
	}
	n, f, isFloat, ok := pyNumber(arg)
	if !ok || (isFloat && (math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f))) {
		return "", pyFormatErr(fmt.Sprintf("%%c requires an int or a unicode character, not %s", pyFormatTypeName(arg)))
	}
	if isFloat {
		n = pyWholeFloat(f)
	}
	if n.Sign() < 0 || n.Cmp(big.NewInt(0x110000)) >= 0 {
		return "", pyFormatErr("%c arg not in range(0x110000)")
	}
	return string(rune(n.Int64())), nil
}

// pyPad pads s with spaces to width runes.
func pyPad(s string, width int, left bool) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	pad := strings.Repeat(" ", width-n)
	if left {
		return s + pad
	}
	return pad + s
}

// pyAssemble joins sign, prefix and digits to width: zeros go between the
// prefix and the digits, spaces outside.
func pyAssemble(sign, prefix, body string, fl pyFlags, width int) string {
	n := len(sign) + len(prefix) + utf8.RuneCountInString(body)
	switch {
	case n >= width:
		return sign + prefix + body
	case fl.left:
		return sign + prefix + body + strings.Repeat(" ", width-n)
	case fl.zero:
		return sign + prefix + strings.Repeat("0", width-n) + body
	}
	return strings.Repeat(" ", width-n) + sign + prefix + body
}

func pySign(neg bool, fl pyFlags) string {
	switch {
	case neg:
		return "-"
	case fl.plus:
		return "+"
	case fl.space:
		return " "
	}
	return ""
}

func pyFormatInt(n *big.Int, c byte, fl pyFlags, width, prec int) string {
	base, prefix := 10, ""
	switch c {
	case 'o':
		base = 8
		if fl.alt {
			prefix = "0o"
		}
	case 'x':
		base = 16
		if fl.alt {
			prefix = "0x"
		}
	case 'X':
		base = 16
		if fl.alt {
			prefix = "0X"
		}
	}
	digits := new(big.Int).Abs(n).Text(base)
	if c == 'X' {
		digits = strings.ToUpper(digits)
	}
	if len(digits) < prec {
		digits = strings.Repeat("0", prec-len(digits)) + digits
	}
	return pyAssemble(pySign(n.Sign() < 0, fl), prefix, digits, fl, width)
}

func pyFormatFloat(f float64, c byte, fl pyFlags, width, prec int) string {
	upper := c == 'E' || c == 'F' || c == 'G'
	sign := pySign(math.Signbit(f) && !math.IsNaN(f), fl)
	if math.IsNaN(f) || math.IsInf(f, 0) {
		s := "inf"
		if math.IsNaN(f) {
			s = "nan"
		}
		if upper {
			s = strings.ToUpper(s)
		}
		return pyAssemble(sign, "", s, fl, width)
	}
	if prec < 0 {
		prec = 6
	}
	a := math.Abs(f)
	var s string
	switch c {
	case 'f', 'F':
		s = strconv.FormatFloat(a, 'f', prec, 64)
		if fl.alt && prec == 0 {
			s += "."
		}
	case 'e', 'E':
		s = strconv.FormatFloat(a, 'e', prec, 64)
		if fl.alt && prec == 0 {
			s = strings.Replace(s, "e", ".e", 1)
		}
	default:
		s = pyFormatG(a, prec, fl.alt)
	}
	if upper {
		s = strings.ToUpper(s)
	}
	return pyAssemble(sign, "", s, fl, width)
}

// pyFormatG is %g: scientific when the exponent is < -4 or >= precision.
func pyFormatG(a float64, prec int, alt bool) string {
	if prec == 0 {
		prec = 1
	}
	e := strconv.FormatFloat(a, 'e', prec-1, 64)
	x, _ := strconv.Atoi(e[strings.IndexByte(e, 'e')+1:])
	var s string
	if x < -4 || x >= prec {
		s = e
	} else {
		s = strconv.FormatFloat(a, 'f', prec-1-x, 64)
	}
	mant, exp := s, ""
	if k := strings.IndexByte(s, 'e'); k >= 0 {
		mant, exp = s[:k], s[k:]
	}
	if alt {
		if !strings.Contains(mant, ".") {
			mant += "."
		}
	} else if strings.Contains(mant, ".") {
		mant = strings.TrimRight(strings.TrimRight(mant, "0"), ".")
	}
	return mant + exp
}

func pyEscapeRune(r rune) string {
	switch {
	case r <= 0xff:
		return fmt.Sprintf(`\x%02x`, r)
	case r <= 0xffff:
		return fmt.Sprintf(`\u%04x`, r)
	}
	return fmt.Sprintf(`\U%08x`, r)
}

// pyRepr is Python's repr of a str.
func pyRepr(s string) string {
	quote := '\''
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	var b strings.Builder
	b.WriteRune(quote)
	for _, r := range s {
		switch {
		case r == quote || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < 0x20 || r == 0x7f:
			b.WriteString(pyEscapeRune(r))
		case r < 0x7f || unicode.IsPrint(r):
			b.WriteRune(r)
		default:
			b.WriteString(pyEscapeRune(r))
		}
	}
	b.WriteRune(quote)
	return b.String()
}

// pyASCII escapes every non-ASCII rune of a repr, as Python's ascii().
func pyASCII(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x80 {
			b.WriteRune(r)
		} else {
			b.WriteString(pyEscapeRune(r))
		}
	}
	return b.String()
}
