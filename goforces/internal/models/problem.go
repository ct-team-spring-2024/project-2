package models

import "time"

type ProblemStatus string

const (
	Draft     ProblemStatus = "Draft"
	Published ProblemStatus = "Published"
	Rejected  ProblemStatus = "Rejected"
)

type Problem struct {
	ProblemId   int           `json:"problemId"`
	OwnerId     int           `json:"ownerId"`
	Title       string        `json:"title"`
	Statement   string        `json:"statement"`
	TimeLimit   int           `json:"timeLimit"`   // in seconds
	MemoryLimit int           `json:"memoryLimit"` // in MB
	Input       string        `json:"input"`
	Output      string        `json:"output"`
	Status      ProblemStatus `json:"status"`
	Feedback    string        `json:"feedback,omitempty"` // In case of rejection
	PublishDate time.Time     `json:"publishDate,omitempty"`
}
