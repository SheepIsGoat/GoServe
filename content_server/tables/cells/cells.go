package cells

type ProfileCell struct {
	Avatar string
	Name   string
	Title  string
}

func (ProfileCell) TemplateName() string {
	return "tableCell/profile"
}

type StatusCell struct {
	Status string
	Color  string
}

func (StatusCell) TemplateName() string {
	return "tableCell/status"
}

var StatusColorMap = map[string]string{
	"Approved": "green",
	"Pending":  "orange",
	"Denied":   "red",
	"Expired":  "grey",
}

var ColorCssMap = map[string]string{
	"green":  "text-green-700  bg-green-100  dark:bg-green-700  dark:text-green-100",
	"orange": "text-orange-700 bg-orange-100 dark:bg-orange-600 dark:text-white",
	"red":    "text-red-700    bg-red-100    dark:bg-red-700    dark:text-red-100",
	"grey":   "text-gray-700   bg-gray-100   dark:text-gray-100 dark:bg-gray-700",
	"purple": "text-white transition-colors bg-purple-600 active:bg-purple-600 hover:bg-purple-700 focus:outline-none focus:shadow-outline-purple",
}

type BasicCell struct {
	Val string
}

func (BasicCell) TemplateName() string {
	return "tableCell/basic"
}

type ModalCell struct {
	LinkText     string
	ModalContent string
}

func (ModalCell) TemplateName() string {
	return "tableCell/modal"
}

type TrashCell struct {
	Filename string
	FileId   string
}

func (TrashCell) TemplateName() string {
	return "tableCell/trash"
}

type HiddenCell struct {
	Val string
}

func (HiddenCell) TemplateName() string {
	return "tableCell/hidden"
}
