package db

import (
	"fmt"
	"sync"
	"time"

	"oj/goforces/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type XMockDB struct {
	submissions      []*models.Submission
	users            []*models.User
	problems         []*models.Problem
	mu               sync.Mutex
	nextSubmissionID int
	nextUserID       int
	nextProblemID    int
}

func NewXMockDB() *XMockDB {

	submissions := []*models.Submission{
		{
			ID:        1,
			UserId:    1,
			ProblemId: 1,
			Code:      "code1",
			Status:    models.Submitted,
		},
	}

	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	users := []*models.User{
		{
			UserId:   1,
			Username: "admin",
			Email:    "admin@example.com",
			Password: string(hashedPass),
			Role:     "admin",
		},
	}

	problems := []*models.Problem{
		{
			ProblemId:   1,
			OwnerId:     1,
			Title:       "Sample Problem",
			Statement:   "Calculate sum of two numbers",
			TimeLimit:   1,
			MemoryLimit: 256,
			Status:      models.Published,
			PublishDate: time.Now(),
		},
	}

	return &XMockDB{
		submissions:      submissions,
		users:            users,
		problems:         problems,
		nextSubmissionID: 2,
		nextUserID:       2,
		nextProblemID:    2,
	}
}

func (m *XMockDB) AddSubmission(s models.Submission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s.ID = m.nextSubmissionID
	m.nextSubmissionID++

	newSubmission := s
	m.submissions = append(m.submissions, &newSubmission)
	return nil
}

func (m *XMockDB) GetUserSubmission(userID int) []models.Submission {
	m.mu.Lock()
	defer m.mu.Unlock()

	var userSubmissions []models.Submission
	for _, submission := range m.submissions {
		if submission.UserId == userID {
			userSubmissions = append(userSubmissions, *submission)
		}
	}
	return userSubmissions
}

func (m *XMockDB) CreateUser(user models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPass)
	user.UserId = m.nextUserID
	m.nextUserID++

	m.users = append(m.users, &user)
	return nil
}

func (m *XMockDB) GetUserByID(userID int) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, user := range m.users {
		if user.UserId == userID {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found") // [[4]]
}

func (m *XMockDB) GetProblemByID(id int) (*models.Problem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.problems {
		if p.ProblemId == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("problem not found")
}

func (m *XMockDB) CreateProblem(problem models.Problem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	problem.ProblemId = m.nextProblemID
	m.nextProblemID++

	m.problems = append(m.problems, &problem)
	return nil
}

func (m *XMockDB) UpdateProblemStatus(problemID int, status models.ProblemStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.problems {
		if p.ProblemId == problemID {
			p.Status = status
			return nil
		}
	}
	return fmt.Errorf("problem not found")
}
