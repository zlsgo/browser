//go:build !nostealth
// +build !nostealth

package browser

import _ "embed"

//go:embed stealth.min.js
var stealth string
