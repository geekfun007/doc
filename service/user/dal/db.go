package dal

import (
	"sync"
	"sync/atomic"
	"time"
)

type UserModel struct {
	ID        int64
	Username  string
	Email     string
	Password  string
	Avatar    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// In-memory store for demonstration; replace with MySQL/PostgreSQL in production.
type UserStore struct {
	mu      sync.RWMutex
	users   map[int64]*UserModel
	byName  map[string]*UserModel
	counter int64
}

var store *UserStore
var once sync.Once

func GetStore() *UserStore {
	once.Do(func() {
		store = &UserStore{
			users:  make(map[int64]*UserModel),
			byName: make(map[string]*UserModel),
		}
	})
	return store
}

func (s *UserStore) Create(username, email, password string) (*UserModel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[username]; exists {
		return nil, ErrUserExists
	}

	id := atomic.AddInt64(&s.counter, 1)
	now := time.Now()
	u := &UserModel{
		ID:        id,
		Username:  username,
		Email:     email,
		Password:  password,
		Avatar:    "",
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.users[id] = u
	s.byName[username] = u
	return u, nil
}

func (s *UserStore) GetByID(id int64) (*UserModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *UserStore) GetByUsername(username string) (*UserModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byName[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *UserStore) Update(id int64, username, email, avatar string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[id]
	if !ok {
		return ErrUserNotFound
	}

	if username != "" && username != u.Username {
		if _, exists := s.byName[username]; exists {
			return ErrUserExists
		}
		delete(s.byName, u.Username)
		u.Username = username
		s.byName[username] = u
	}
	if email != "" {
		u.Email = email
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	u.UpdatedAt = time.Now()
	return nil
}
