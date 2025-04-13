package db

import (
	"sync"

	"oj/goforces/internal/models"
)

type MockDB struct {
	submissions []*models.Submission
	mu          sync.Mutex
}

func NewMockDB() *MockDB {
	submissions := make([]*models.Submission, 0)
	submissions = append(submissions, models.NewSubmission(1, 1))
	submissions = append(submissions, models.NewSubmission(1, 2))
	return &MockDB{
		submissions: submissions,
	}
}

func (m *MockDB) GetUserSubmission(userID int) []models.Submission {
	m.mu.Lock()

	var userSubmissions []models.Submission

	for _, submission := range m.submissions {
		if submission.UserId == userID {
			userSubmissions = append(userSubmissions, *submission)
		}
	}

	m.mu.Unlock()
	return userSubmissions
}

func (m *MockDB) AddSubmission(s models.Submission) error {
	m.mu.Lock()

	m.submissions = append(m.submissions, &s)

	m.mu.Unlock()
	return nil
}
