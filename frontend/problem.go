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
	Submissions      []Submission
	CurrentPage      int
	Limit            int
	HasNextPage      bool
	TotalPages       int
	Username         string
	Email            string
	MemberSince      string
	TotalSubmissions int
	SolvedProblems   int
	SolveRate        int
}

type ProblemsPageData struct {
	Page           string
	ClientUsername string
	IsClientAdmin  bool
	Problems       []Problem
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
