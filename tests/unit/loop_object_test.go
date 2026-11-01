package miya_test

import (
	"fmt"
	"sync"
	"testing"

	miya "github.com/zipreport/miya"
)

func TestLoopObjectNotReused(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	cases := []struct{ name, src, want string }{
		{"second_loop", `{% set ns = namespace(l=none) %}{% for i in [1,2,3] %}{% set ns.l = loop %}{% endfor %}{% for j in [7,8] %}{% endfor %}{{ ns.l.index }}`, "3"},
		{"single_loop", `{% set ns = namespace(l=none) %}{% for i in [1,2,3] %}{% set ns.l = loop %}{% endfor %}{{ ns.l.index }}|{{ ns.l.last }}`, "3|true"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := env.RenderString(tc.src, miya.NewContext())
			if err != nil || got != tc.want {
				t.Fatalf("got %q, err %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestLoopObjectConcurrent(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	tmpl, err := env.FromString(`{% set ns = namespace(l=none) %}{% for i in xs %}{% set ns.l = loop %}{% endfor %}{% for j in ys %}{{ j }}{% endfor %}{{ ns.l.index }}`)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			a, b := 10*g+1, 10*g+2
			want := fmt.Sprintf("%d%d3", a, b)
			for i := 0; i < 50; i++ {
				ctx := miya.NewContextFrom(map[string]interface{}{
					"xs": []interface{}{1, 2, 3},
					"ys": []interface{}{a, b},
				})
				got, err := tmpl.Render(ctx)
				if err != nil || got != want {
					t.Errorf("goroutine %d: got %q, err %v; want %q", g, got, err, want)
					continue
				}
			}
		}(g)
	}
	wg.Wait()
}
