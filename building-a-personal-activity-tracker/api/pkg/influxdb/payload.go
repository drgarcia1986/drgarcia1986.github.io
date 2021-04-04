package influxdb

import (
	"bytes"
	"fmt"
	"strings"
)

type category struct {
	name  string
	score int
}

type metric struct {
	category category
	app      string
}

var (
	WORK_CATEGORY    = category{name: "work", score: 1}
	SOCIAL_CATEGORY  = category{name: "social", score: 0}
	GAME_CATEGORY    = category{name: "game", score: -1}
	UNKNOWN_CATEGORY = category{name: "unknown", score: 0}
)

func convertTerminalAppName(title string) string {
	windowTitle := strings.Split(title, " - ")
	return windowTitle[1]
}

func convertBrowserCategory(title string) category {
	if strings.Contains(title, "Twitter") {
		return SOCIAL_CATEGORY
	}
	return WORK_CATEGORY
}

func convertToMetric(class, title string) metric {
	switch class {
	case "gnome-terminal-server":
		return metric{
			category: WORK_CATEGORY,
			app:      convertTerminalAppName(title),
		}
	case "Navigator":
		return metric{
			category: convertBrowserCategory(title),
			app:      "FireFox",
		}
	case "telegram":
		return metric{
			category: SOCIAL_CATEGORY,
			app:      class,
		}
	case "Steam":
		return metric{
			category: GAME_CATEGORY,
			app:      "Steam",
		}
	case "WM_CLASS: not found.":
		if title == "UNDERTALE" {
			return metric{
				category: GAME_CATEGORY,
				app:      title,
			}
		}
		return metric{category: UNKNOWN_CATEGORY, app: title}
	default:
		return metric{category: UNKNOWN_CATEGORY, app: title}
	}
}

func buildPayload(class, title string) *bytes.Buffer {
	m := convertToMetric(class, title)
	return bytes.NewBufferString(
		fmt.Sprintf(
			"productivity,category=\"%s\",app=\"%s\" value=%d",
			m.category.name,
			m.app,
			m.category.score,
		),
	)
}
