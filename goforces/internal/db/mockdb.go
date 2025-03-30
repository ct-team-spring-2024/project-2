package db

import (
	"sync"

	"oj/goforces/internal"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)


type MockDB struct {
	submissions []*internal.Submission
	mu          sync.Mutex
}

func NewMockDB() *MockDB {
	submissions := make([]*internal.Submission, 0)
	submissions = append(submissions, internal.NewSubmission(1, 1))
	submissions = append(submissions, internal.NewSubmission(1, 2))
	return &MockDB{
		submissions: submissions,
	}
}

func (m *MockDB) GetUserSubmission(user internal.User) []internal.Submission {
	m.mu.Lock()

	logrus.Infof("Current => %s", spew.Sdump(m.submissions))
	logrus.Infof("user => %+v", user)
	var userSubmissions []internal.Submission

	for _, submission := range m.submissions {
		if submission.UserId == user.UserId {
			userSubmissions = append(userSubmissions, *submission)
		}
	}

	m.mu.Unlock()
	return userSubmissions
}
