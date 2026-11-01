package miya_test

import (
	"fmt"
	"testing"

	miya "github.com/zipreport/miya"
)

func benchRender(b *testing.B, src string) {
	env := miya.NewEnvironment()
	items := make([]map[string]interface{}, 1000)
	for i := range items {
		items[i] = map[string]interface{}{"text": fmt.Sprintf("  Sample Text Item Number %d With Extra Spacing  ", i+1)}
	}
	tmpl, err := env.FromString(src)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := miya.NewContext()
		ctx.Set("items", items)
		if _, err := tmpl.Render(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderLoop(b *testing.B) {
	benchRender(b, `{% for item in items %}{{ item.text }}{% endfor %}`)
}

func BenchmarkRenderFilterChain(b *testing.B) {
	benchRender(b, `{% for item in items %}{{ item.text|trim|lower|replace(" ", "_")|upper|reverse|truncate(20)|center(25) }}{% endfor %}`)
}
