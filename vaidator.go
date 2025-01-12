package browser

import (
	"strings"

	"github.com/go-rod/rod"
	"github.com/sohaha/zlsgo/zstring"
)

type filterRules struct {
	selector string
	value    string
	attr     string
	isEq     bool
}

func filterElementsRules(f ...string) []filterRules {
	var rules []filterRules
	for _, v := range f {
		r := filterRules{}
		s := strings.SplitN(v, "!=", 2)
		if len(s) != 2 {
			s = strings.SplitN(v, "=", 2)
			r.isEq = true
		}
		if len(s) < 2 {
			continue
		}

		r.value = strings.TrimSpace(s[1])

		s = strings.SplitN(s[0], ",", 2)
		r.selector = strings.TrimSpace(s[0])
		if len(s) > 1 {
			r.attr = strings.TrimSpace(s[len(s)-1])
		}
		rules = append(rules, r)
	}

	return rules
}

func filterElements(f ...string) func(e *rod.Element) bool {
	var rules []filterRules
	if len(f) > 0 {
		rules = filterElementsRules(f...)
	}

	return func(e *rod.Element) bool {
		if len(rules) == 0 {
			return true
		}
		for _, v := range rules {
			ele, err := e.Element(v.selector)
			if err != nil {
				return false
			}

			var match bool
			if v.attr != "" {
				p, err := ele.Property(v.attr)
				if err != nil {
					return false
				}
				match = zstring.Match(p.String(), v.value, true)
			} else {
				match = zstring.Match(ele.MustText(), v.value, true)
			}

			return (v.isEq && match) || (!v.isEq && !match)
		}
		return true
	}
}
