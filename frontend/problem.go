package frontend

type Problem struct {
	Id          int
	Statement   string
	Title       string
	ProblemName string
	TimeLimit   string
	MemoryLimit string
}

type Submission struct {
	ProblemName    string
	Status         string
	SubmissionDate string
}

type ProfilePageData struct {
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
	Problems    []Problem
	CurrentPage int
	Limit       int
	HasNextPage bool
	TotalPages  int
}
