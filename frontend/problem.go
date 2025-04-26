package frontend

type Problem struct {
	Id          int
	Statement   string
	Title       string
	ProblemName string
	TimeLimit   int
	MemoryLimit int
	Status      string
}

type Submission struct {
	ProblemName    string
	Status         string
	SubmissionDate string
}

type ProfilePageData struct {
	Page             string
	ClientUsername   string
	IsClientAdmin    bool
	IsUserAdmin      bool
	Username         string
	Submissions      []Submission
	Email            string
	MemberSince      string
	TotalSubmissions int
	SolvedProblems   int
	SolveRate        int
}

type ProblemPageProblemSummary struct {
	Id    int
	Title string
}

type ProblemsPageData struct {
	Page           string
	ClientUsername string
	IsClientAdmin  bool
	Problems       []ProblemPageProblemSummary
	CurrentPage    int
	Limit          int
	HasNextPage    bool
	TotalPages     int
}

type AddProblemPageData struct {
	Page           string
	ClientUsername string
	IsClientAdmin  bool
}

type SubmitPageData struct {
	Page               string
	ClientUsername     string
	IsClientAdmin      bool
	ProblemId          int
}

type SingleProblemPageData struct {
	Page               string
	ClientUsername     string
	IsClientAdmin      bool
	ProblemId          int
	ProblemStatement   string
	ProblemTitle       string
	ProblemTimeLimit   int
	ProblemMemoryLimit int
}

type ProblemSummary struct {
	Id     string
	Title  string
	Status string
}

type MyProblemsPageData struct {
	Page           string
	ClientUsername string
	IsClientAdmin  bool
	Problems       []ProblemSummary
}

type EditProblemPageData struct {
	Page           string
	ClientUsername string
	IsClientAdmin  bool
	Problem        Problem
}

type ManageProblemsPageData struct {
	Page           string
	ClientUsername string
	IsClientAdmin  bool
	Problems       []ProblemSummary
}


type TestsStatus map[string]struct{
	Status string
}

type SubmissionsPageEntryData struct {
	Id int
	ProblemId int
	TestsStatus TestsStatus
	SubmissionStatus string
	Score int
}

type SubmissionsPageData struct {
	Page           string
	ClientUsername string
	IsClientAdmin  bool
	Submissions    []SubmissionsPageEntryData
}
