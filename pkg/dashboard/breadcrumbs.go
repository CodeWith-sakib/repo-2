package dashboard

import "fmt"

type Breadcrumb struct {
	Label string
	URL   string
}

type BreadcrumbTrail struct {
	crumbs []Breadcrumb
}

func NewBreadcrumbTrail() *BreadcrumbTrail {
	return &BreadcrumbTrail{
		crumbs: []Breadcrumb{
			{Label: "Home", URL: "/"},
		},
	}
}

func (b *BreadcrumbTrail) Add(label, url string) {
	b.crumbs = append(b.crumbs, Breadcrumb{Label: label, URL: url})
}

func (b *BreadcrumbTrail) RenderHTML() string {
	var html string
	for i, c := range b.crumbs {
		if i > 0 {
			html += " > "
		}
		html += fmt.Sprintf("<a href='%s'>%s</a>", c.URL, c.Label)
	}
	return html
}
