package api


type EvalCodeRequest struct {
	Code    string   `json:"code"`
	Inputs  []string `json:"inputs"`
	Timelimit   int      `json:"timelimit"`
	Memorylimit int      `json:"memorylimit"`
}
