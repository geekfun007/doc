package dal

import (
	"sync"
	"sync/atomic"
	"time"
)

type ArticleModel struct {
	ID         int64
	Title      string
	Content    string
	AuthorID   int64
	AuthorName string
	Status     int32 // 0=draft, 1=published, 2=deleted
	ViewCount  int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// In-memory store for demonstration; replace with MySQL/PostgreSQL in production.
type ArticleStore struct {
	mu       sync.RWMutex
	articles map[int64]*ArticleModel
	counter  int64
}

var store *ArticleStore
var once sync.Once

func GetStore() *ArticleStore {
	once.Do(func() {
		store = &ArticleStore{
			articles: make(map[int64]*ArticleModel),
		}
	})
	return store
}

func (s *ArticleStore) Create(title, content string, authorID int64, authorName string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := atomic.AddInt64(&s.counter, 1)
	now := time.Now()
	s.articles[id] = &ArticleModel{
		ID:         id,
		Title:      title,
		Content:    content,
		AuthorID:   authorID,
		AuthorName: authorName,
		Status:     1,
		ViewCount:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return id, nil
}

func (s *ArticleStore) GetByID(id int64) (*ArticleModel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.articles[id]
	if !ok || a.Status == 2 {
		return nil, ErrArticleNotFound
	}
	a.ViewCount++
	return a, nil
}

func (s *ArticleStore) ListByAuthor(authorID int64, page, pageSize int32) ([]*ArticleModel, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var all []*ArticleModel
	for _, a := range s.articles {
		if a.Status != 2 && (authorID == 0 || a.AuthorID == authorID) {
			all = append(all, a)
		}
	}

	total := int64(len(all))
	start := int((page - 1) * pageSize)
	if start >= len(all) {
		return nil, total
	}
	end := start + int(pageSize)
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total
}

func (s *ArticleStore) Update(id int64, title, content string, status int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.articles[id]
	if !ok {
		return ErrArticleNotFound
	}
	if title != "" {
		a.Title = title
	}
	if content != "" {
		a.Content = content
	}
	if status >= 0 {
		a.Status = status
	}
	a.UpdatedAt = time.Now()
	return nil
}

func (s *ArticleStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.articles[id]
	if !ok {
		return ErrArticleNotFound
	}
	a.Status = 2
	a.UpdatedAt = time.Now()
	return nil
}
