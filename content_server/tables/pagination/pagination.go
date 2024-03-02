package pagination

import (
	"fmt"
)

type Pagination struct {
	Data    PaginData
	Config  PaginConfig
	Display PaginDisplay
}

type PaginData struct {
	TableName string
	ItemTotal uint32
}
type PaginConfig struct {
	CurrentPage  uint32
	ItemsPerPage uint32
}
type PaginDisplay struct {
	TotalPages uint32
	PageList   []string
	ItemStart  uint32
	ItemEnd    uint32
}

func (p *Pagination) _buildPageList(innerPageRange uint32) {
	p.Display.PageList = []string{"1"}
	if p.Display.TotalPages <= 1 {
		// 1
		return
	}
	// must display between 5-total pages
	innerPageRange = min(innerPageRange, (p.Display.TotalPages-1)/2)
	innerPageRange = max(innerPageRange, 2)

	if p.Config.CurrentPage > 1+innerPageRange && p.Display.TotalPages > innerPageRange*2+1 {
		// -  -  - - x - - - -
		// 1 ... 4 5 6 7 8 9 10
		//    ^
		p.Display.PageList = append(p.Display.PageList, "...")
	} else {
		// - - - x - - - -
		// 1 2 3 4 5 6 7 8
		//   ^
		p.Display.PageList = append(p.Display.PageList, "2")
	}
	if p.Display.TotalPages == 2 {
		// 1 2
		return
	}

	leftPage := max(min(p.Config.CurrentPage-min(innerPageRange, p.Config.CurrentPage)+2, p.Display.TotalPages-(2*innerPageRange)+2), 3)

	// -  -  - - x - - -   -
	// 1 ... 4 5 6 7 8 ... 99
	//               ^
	// rightPage := min(max(p.CurrentPage+innerPageRange-2, 3), p.TotalPages) // 3 <= rightPage <= TotalPages
	rightPage := min(max(p.Config.CurrentPage+2*innerPageRange-min(innerPageRange, p.Config.CurrentPage)-2, 3), p.Display.TotalPages)

	for page := leftPage; page <= rightPage; page++ {
		// -  -  - - x - - -   -
		// 1 ... 4 5 6 7 8 ... 99
		//       ^ ^ ^ ^ ^
		p.Display.PageList = append(p.Display.PageList, fmt.Sprint(page))
	}

	if rightPage == p.Display.TotalPages {
		// -  -  - - x - -
		// 1 ... 4 5 6 7 8
		//       ^ ^ ^ ^ ^
		return
	}

	if rightPage+1 < p.Display.TotalPages-1 {
		// -  -  - - x - - -   -
		// 1 ... 4 5 6 7 8 ... 99
		//                 ^
		p.Display.PageList = append(p.Display.PageList, "...")
	} else if rightPage+1 == p.Display.TotalPages-1 {
		// -  -  - - x - - - -
		// 1 ... 4 5 6 7 8 9 10
		//                 ^
		p.Display.PageList = append(p.Display.PageList, fmt.Sprint(rightPage+1))
	}
	// -  -  - - x - - -   -
	// 1 ... 4 5 6 7 8 ... 99
	//                     ^
	p.Display.PageList = append(p.Display.PageList, fmt.Sprint(p.Display.TotalPages))
}

func (p *Pagination) _setPageBounds(innerPageRange uint32) {
	p.Config.ItemsPerPage = max(p.Config.ItemsPerPage, 1)
	p.Display.TotalPages = (p.Data.ItemTotal + p.Config.ItemsPerPage - 1) / p.Config.ItemsPerPage
	p.Config.CurrentPage = min(p.Config.CurrentPage, p.Display.TotalPages)
	if p.Config.CurrentPage == 0 {
		p.Config.CurrentPage = 1
	}
	fmt.Printf("Setting currentPage to %v. Total Pages %v\n", p.Config.CurrentPage, p.Display.TotalPages)
	p.Display.ItemStart = (p.Config.CurrentPage-1)*p.Config.ItemsPerPage + 1
	p.Display.ItemEnd = min(p.Display.ItemStart+p.Config.ItemsPerPage-1, p.Data.ItemTotal)
}

func (p *Pagination) Init(
	TableName string,
	ItemTotal uint32,
	CurrentPage uint32,
	ItemsPerPage uint32,
	innerPageRange uint32,
) {
	p.Data = PaginData{
		TableName: TableName,
		ItemTotal: ItemTotal,
	}

	p.Config = PaginConfig{
		CurrentPage:  CurrentPage,
		ItemsPerPage: ItemsPerPage,
	}

	p._setPageBounds(innerPageRange)
	p._buildPageList(innerPageRange)
}
