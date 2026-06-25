package specs

import "strings"

type Spec struct {
	Display, Processor, RAM, Battery, Camera, Price string
}

var catalog = map[string]Spec{
	"iphone 16": {
		Display:   "6.1-inch Super Retina XDR OLED",
		Processor: "A18 Bionic",
		RAM:       "8GB",
		Battery:   "3561 mAh",
		Camera:    "48MP main + 12MP ultrawide",
		Price:     "from $799",
	},
	"iphone 15": {
		Display:   "6.1-inch Super Retina XDR OLED",
		Processor: "A16 Bionic",
		RAM:       "6GB",
		Battery:   "3349 mAh",
		Camera:    "48MP main + 12MP ultrawide",
		Price:     "from $699",
	},
}

func Lookup(product string) (Spec, bool) {
	s, ok := catalog[strings.ToLower(product)]
	return s, ok
}
