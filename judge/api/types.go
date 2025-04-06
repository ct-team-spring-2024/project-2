package api


type EvalCodeRequest struct {
	Code    string   `json:"code"`
	Inputs  []string `json:"inputs"`
	Timeout int      `json:"timeout"`
}
