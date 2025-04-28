package services

import (
	"errors"
	"sort"
	"sync"
	"time"

	"oj/goforces/internal/db"
	"oj/goforces/internal/models"

	"github.com/sirupsen/logrus"
)

var (
	problemsMutex sync.Mutex
)

func CreateProblem(problem models.Problem) (models.Problem, error) {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	//problem.ProblemId = problemIdCounter
	//	problemIdCounter++
	//problemsStore = append(problemsStore, problem)
	err := db.DB.CreateProblem(problem)
	if err != nil {
		return problem, err
	}

	return problem, nil
}

func UpdateProblem(problemId int, updatedProblem models.Problem) (models.Problem, error) {
	// TODO: This caused deadlock. Why are using these mutexes ????????
	// problemsMutex.Lock()
	// defer problemsMutex.Unlock()
	db.DB.UpdateProblem(problemId, updatedProblem)
	return models.Problem{}, errors.New("problem not found")
}

func GetMyProblems(ownerId int) []models.Problem {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	var userProblems []models.Problem
	problems, _ := db.DB.GetProblems()
	for _, p := range problems {
		if p.OwnerId == ownerId {
			userProblems = append(userProblems, p)
		}
	}
	return userProblems
}

func GetProblemById(problemId int) (models.Problem, error) {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	problems, _ := db.DB.GetProblems()

	for _, p := range problems {
		if p.ProblemId == problemId {
			return p, nil
		}
	}
	return models.Problem{}, errors.New("problem not found")
}

func GetPublishedProblems(page int, pageSize int) []models.Problem {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	var published []models.Problem
	problems, _ := db.DB.GetProblems()
	logrus.Info("came heredjfsdkjfdslkjdsflkdf")
	for _, p := range problems {
		logrus.Infof("PP => %v", p)
		logrus.Info("cameHere")
		if p.Status == "Published" {
			published = append(published, p)
		}
	}

	sort.Slice(published, func(i, j int) bool {
		return published[i].PublishDate.After(published[j].PublishDate)
	})

	start := (page - 1) * pageSize
	if start >= len(published) {
		return []models.Problem{}
	}
	end := start + pageSize
	if end > len(published) {
		end = len(published)
	}
	return published[start:end]
}

func GetAllProblems() []models.Problem {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()
	problems, _ := db.DB.GetProblems()
	return problems
}

func UpdateProblemStatus(problemId int, newStatus string, feedback string) (models.Problem, error) {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()
	problems, _ := db.DB.GetProblems()

	for _, p := range problems {
		if p.ProblemId == problemId {
			if newStatus != string(models.Published) && newStatus != string(models.Draft) && newStatus != string(models.Rejected) {
				return models.Problem{}, errors.New("invalid status")
			}
			p.Status = models.ProblemStatus(newStatus)
			if newStatus == string(models.Published) {
				p.PublishDate = time.Now()
				p.Feedback = ""
			} else if newStatus == string(models.Draft) {
				p.PublishDate = time.Time{}
				p.Feedback = ""
			} else if newStatus == string(models.Rejected) {
				p.PublishDate = time.Time{}
				p.Feedback = feedback
			}

			UpdateProblem(p.ProblemId, p)
			return p, nil
		}
	}
	return models.Problem{}, errors.New("problem not found")
}
