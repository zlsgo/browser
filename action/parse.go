package action

import (
	"time"

	"github.com/sohaha/zlsgo/zjson"
)

func parseAction(actionArray []*zjson.Res) (actions Actions) {
	if len(actionArray) == 0 {
		return nil
	}

	for _, v := range actionArray {
		actionType := v.Get("action").String()
		name := v.Get("name").String()
		value := v.Get("value").String()
		timeout := v.Get("timeout").Int()
		selector := v.Get("selector").String()
		next := v.Get("next").Array()
		// vaidator := v.Get("vaidator").Value()
		nextActions := parseAction(next)
		action := Action{
			Name:     name,
			Next:     nextActions,
			Vaidator: nil,
		}
		switch actionType {
		case "WaitDOMStable":
			action.Action = WaitDOMStable(0.5, time.Second*time.Duration(timeout))
		case "InputEnter":
			action.Action = InputEnter(selector, value)
		case "Elements":
			action.Action = Elements(selector)
		case "Screenshot":
			action.Action = Screenshot("")
		case "ClickNewPage":
			action.Action = ClickNewPage(selector)
		case "ActivatePage":
			action.Action = ActivatePage()
		case "ClosePage":
			action.Action = ClosePage()
		}
		actions = append(actions, action)
	}
	return
}
