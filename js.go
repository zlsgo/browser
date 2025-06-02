package browser

import (
	_ "embed"

	"github.com/ysmood/gson"
)

// npx extract-stealth-evasions

var jsWaitDOMContentLoad = `()=>{const n=this===window;return new Promise((e,t)=>{if(n){if("complete"===document.readyState)return e();window.addEventListener("DOMContentLoaded",e)}else void 0===this.complete||this.complete?e():(this.addEventListener("DOMContentLoaded",e),this.addEventListener("error",t))})}`

var jsWaitLoad = `()=>{const n=this===window;return new Promise((e,t)=>{if(n){if("complete"===document.readyState)return e();window.addEventListener("load",e)}else void 0===this.complete||this.complete?e():(this.addEventListener("load",e),this.addEventListener("error",t))})}`

func (page *Page) EvalJS(js string, params ...interface{}) (gson.JSON, error) {
	resp, err := page.Timeout().page.Eval(js, params...)
	if err != nil {
		return gson.JSON{}, err
	}
	return resp.Value, nil
}
