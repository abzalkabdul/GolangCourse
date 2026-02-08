package models

import (
	"errors"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[int]*Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}
}

func (s *TaskStore) Create(title string) (*Task, error) {
	if title == "" {
		return nil, errors.New("invalid title")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[s.nextID] = task
	s.nextID++

	return task, nil
}

func (s *TaskStore) GetByID(id int) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

func (s *TaskStore) GetAll() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

func (s *TaskStore) Update(id int, done bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return errors.New("task not found")
	}

	task.Done = done
	return nil
}

func (s *TaskStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return errors.New("task not found")
	}

	delete(s.tasks, id)
	return nil
}

func (s *TaskStore) GetFiltered(done *bool) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range s.tasks {
		if done == nil || task.Done == *done {
			tasks = append(tasks, task)
		}
	}

	return tasks
}
