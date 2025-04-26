package services

import (
	"errors"
	"sort"
	"sync"
	"time"

	"oj/goforces/internal/models"

	"github.com/sirupsen/logrus"
)

var (
	problemsStore    []models.Problem
	problemsMutex    sync.Mutex
	problemIdCounter = 1
)

func CreateProblem(problem models.Problem) (models.Problem, error) {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	problem.ProblemId = problemIdCounter
	problemIdCounter++
	problemsStore = append(problemsStore, problem)
	return problem, nil
}

func UpdateProblem(ownerId int, updatedProblem models.Problem) (models.Problem, error) {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	for i, p := range problemsStore {
		if p.ProblemId == updatedProblem.ProblemId {
			if p.OwnerId != ownerId {
				return models.Problem{}, errors.New("unauthorized: not the owner")
			}
			p.Title = updatedProblem.Title
			p.Statement = updatedProblem.Statement
			p.TimeLimit = updatedProblem.TimeLimit
			p.MemoryLimit = updatedProblem.MemoryLimit
			p.Inputs = updatedProblem.Inputs
			p.Outputs = updatedProblem.Outputs
			problemsStore[i] = p
			return p, nil
		}
	}
	return models.Problem{}, errors.New("problem not found")
}

func GetMyProblems(ownerId int) []models.Problem {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	var userProblems []models.Problem
	for _, p := range problemsStore {
		logrus.Info(p)
		logrus.Info(p.OwnerId)
		logrus.Info(ownerId)
		if p.OwnerId == ownerId {
			logrus.Info(p)
			userProblems = append(userProblems, p)
		}
	}
	return userProblems
}

func GetProblemById(problemId int) (models.Problem, error) {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	for _, p := range problemsStore {
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
	for _, p := range problemsStore {
		logrus.Infof("PP => %v", p)
		if p.Status == "published" {
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
	return problemsStore
}

func UpdateProblemStatus(problemId int, newStatus string, feedback string) (models.Problem, error) {
	problemsMutex.Lock()
	defer problemsMutex.Unlock()

	for i, p := range problemsStore {
		if p.ProblemId == problemId {
			if newStatus != "draft" && newStatus != "published" && newStatus != "rejected" {
				return models.Problem{}, errors.New("invalid status")
			}
			p.Status = models.ProblemStatus(newStatus)
			if newStatus == "published" {
				p.PublishDate = time.Now()
				p.Feedback = ""
			} else if newStatus == "draft" {
				p.PublishDate = time.Time{}
				p.Feedback = ""
			} else if newStatus == "rejected" {
				p.PublishDate = time.Time{}
				p.Feedback = feedback
			}
			problemsStore[i] = p
			return p, nil
		}
	}
	return models.Problem{}, errors.New("problem not found")
}
