package db

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"oj/goforces/internal/models"
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
	return &XMockDB{
		submissions:      make([]*models.Submission, 0, 0),
		users:            make([]*models.User, 0, 0),
		problems:         make([]*models.Problem, 0, 0),
		nextSubmissionID: 0,
		nextUserID:       0,
		nextProblemID:    0,
	}
}

func (m *XMockDB) AddSubmission(s models.Submission) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s.ID = m.nextSubmissionID
	m.nextSubmissionID++

	newSubmission := s
	m.submissions = append(m.submissions, &newSubmission)
	return s.ID, nil
}


func (m *XMockDB) UpdateSubmissionStatus(submissionID int, status models.SubmissionStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	logrus.Infof("Updateing. s => %d", submissionID)
	logrus.Infof("Updateing. status => %+v", status)
	logrus.Infof("Updateing. m => %+v", m.submissions)
	for _, submission := range m.submissions {
		if submissionID == submission.ID {
			submission.SubmissionStatus = status
		}
	}
}

func (m *XMockDB) UpdateTestStatus(s models.Submission, testId string, testStatus models.TestStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, submission := range m.submissions {
		if submission.ID == s.ID {
			submission.TestsStatus[testId] = testStatus
		}
	}
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


func (m *XMockDB) GetSubmission(submissionID int) models.Submission {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, submission := range m.submissions {
		if submission.ID == submissionID {
			return *submission
		}
	}
	logrus.Fatal("No Submission Found")
	return models.Submission{}
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
	return nil, fmt.Errorf("user not found")
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

// Not added to mock database
func (m *XMockDB) GetProblems() ([]models.Problem, error) {
	return make([]models.Problem, 0), nil
}
func (m *XMockDB) UpdateProblem(problemId int, newProblem models.Problem) error {
	return nil
}
func (m *XMockDB) GetUsers() []models.User {
	return make([]models.User, 0)

}
func (m *XMockDB) UpdateUsers(userId int, newUser models.User) error {
	return nil
}
