package charts

var HexColors = map[string]string{
	"teal":   "#0694a2",
	"blue":   "#1c64f2",
	"violet": "#7e3af2",
}

type HexColor struct {
	Name    string
	HexCode string
}

// color palettes should be accessible via WCAG
// https://venngage.com/tools/accessible-color-palette-generator
type ColorPalette []HexColor

var ColorPalettes = map[string]ColorPalette{
	"Dark-To-Light": {
		{
			Name:    "green",
			HexCode: "#029356",
		},
		{
			Name:    "teal",
			HexCode: "#009EB0",
		},
		{
			Name:    "blue",
			HexCode: "#0073E6",
		},
		{
			Name:    "indigo",
			HexCode: "#606FF3",
		},
		{
			Name:    "violet",
			HexCode: "#9B8BF4",
		},
	},
	"Contrasting-1": {
		{
			Name:    "dark-orange",
			HexCode: "#C44601",
		},
		{
			Name:    "orange",
			HexCode: "#F57600",
		},
		{
			Name:    "baby-blue",
			HexCode: "#8BABF1",
		},
		{
			Name:    "blue",
			HexCode: "#0073E6",
		},
		{
			Name:    "dark-blue",
			HexCode: "#054FB9",
		},
	},
}
