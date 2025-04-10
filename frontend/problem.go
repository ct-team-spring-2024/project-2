package frontend

type Problem struct {
	Id        int
	Statement string
	Title     string
}
type PageData struct {
	Problems    []Problem
	CurrentPage int
	Limit       int
	HasNextPage bool
	TotalPages  int
}
