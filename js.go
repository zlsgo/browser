package browser

import (
	_ "embed"
)

//go:embed stealth.min.js
var stealth string

var jsWaitDOMContentLoad = `()=>{const n=this===window;return new Promise((e,t)=>{if(n){if("complete"===document.readyState)return e();window.addEventListener("DOMContentLoaded",e)}else void 0===this.complete||this.complete?e():(this.addEventListener("DOMContentLoaded",e),this.addEventListener("error",t))})}`

var jsWaitLoad = `()=>{const n=this===window;return new Promise((e,t)=>{if(n){if("complete"===document.readyState)return e();window.addEventListener("load",e)}else void 0===this.complete||this.complete?e():(this.addEventListener("load",e),this.addEventListener("error",t))})}`
