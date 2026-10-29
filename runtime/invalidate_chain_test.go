package runtime

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

// Issue #4: InvalidateTemplate drops an entry by its own key even when its
// hierarchy records no chain, and drops entries whose chain contains the name.
func TestIssue4InvalidateTemplateByKeyAndChain(t *testing.T) {
	cache := NewInheritanceCache()
	cache.StoreHierarchy("bare.html", &InheritanceHierarchy{})
	cache.StoreHierarchy("child.html", &InheritanceHierarchy{TemplateMap: map[string]*parser.TemplateNode{
		"child.html": {}, "base.html": {},
	}})
	cache.StoreHierarchy("other.html", &InheritanceHierarchy{TemplateMap: map[string]*parser.TemplateNode{
		"other.html": {},
	}})

	cache.InvalidateTemplate("bare.html")
	if _, found := cache.GetHierarchy("bare.html"); found {
		t.Error("bare.html: entry with no chain survived InvalidateTemplate of its own name")
	}

	cache.InvalidateTemplate("base.html")
	if _, found := cache.GetHierarchy("child.html"); found {
		t.Error("child.html: entry whose chain contains base.html survived InvalidateTemplate(base.html)")
	}
	if _, found := cache.GetHierarchy("other.html"); !found {
		t.Error("other.html: unrelated entry dropped by InvalidateTemplate(base.html)")
	}
}
