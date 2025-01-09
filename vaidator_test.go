package browser

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func Test_filterElementsRules(t *testing.T) {
	tt := zlsgo.NewTest(t)

	rules := filterElementsRules("a,href=xxx.com", "a h2 != xxx.com ")
	tt.Equal(len(rules), 2)

	for i, v := range rules {
		switch i {
		case 0:
			tt.Equal(v.selector, "a", true)
			tt.Equal(v.attr, "href", true)
			tt.Equal(v.isEq, true, true)
			tt.Equal(v.value, "xxx.com", true)
		case 1:
			tt.Equal(v.selector, "a h2", true)
			tt.Equal(v.attr, "", true)
			tt.Equal(v.isEq, false, true)
			tt.Equal(v.value, "xxx.com", true)
		}
	}
}
