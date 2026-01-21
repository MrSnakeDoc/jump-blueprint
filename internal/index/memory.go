package index

import (
	"sync"
	"time"

	"github.com/MrSnakeDoc/jump/internal/domain"
)

// MemoryIndex provides in-memory storage and lookup for services and bookmarks
// It acts as a fallback when Redis is unavailable
type MemoryIndex struct {
	mu                 sync.RWMutex
	services           map[string]*domain.Service  // ID -> Service
	bookmarks          map[string]*domain.Bookmark // ID -> Bookmark
	lastReload         time.Time                   // Timestamp of last services reload
	lastBookmarkReload time.Time                   // Timestamp of last bookmarks reload
}

// NewMemoryIndex creates a new memory index
func NewMemoryIndex() *MemoryIndex {
	return &MemoryIndex{
		services:  make(map[string]*domain.Service),
		bookmarks: make(map[string]*domain.Bookmark),
	}
}

// UpdateServices replaces all services in the index
func (idx *MemoryIndex) UpdateServices(services []*domain.Service) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear and rebuild
	idx.services = make(map[string]*domain.Service, len(services))
	for _, service := range services {
		idx.services[service.ID] = service
	}
	idx.lastReload = time.Now()
}

// GetService retrieves a service by ID
func (idx *MemoryIndex) GetService(id string) (*domain.Service, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	service, ok := idx.services[id]
	return service, ok
}

// GetAllServices returns all services
func (idx *MemoryIndex) GetAllServices() []*domain.Service {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	services := make([]*domain.Service, 0, len(idx.services))
	for _, service := range idx.services {
		services = append(services, service)
	}
	return services
}

// AddService adds or updates a single service
func (idx *MemoryIndex) AddService(service *domain.Service) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.services[service.ID] = service
}

// DeleteService removes a service from the index
func (idx *MemoryIndex) DeleteService(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.services, id)
}

// Count returns the number of services in the index
func (idx *MemoryIndex) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.services)
}

// IncrementCounter increments the usage counter for a service
func (idx *MemoryIndex) IncrementCounter(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if service, ok := idx.services[id]; ok {
		service.Counter++
	}
}

// GetLastReload returns the timestamp of the last services reload
func (idx *MemoryIndex) GetLastReload() time.Time {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.lastReload
}

// ─────────────────────────────────────────────────────────────────
// Bookmark methods
// ─────────────────────────────────────────────────────────────────

// UpdateBookmarks replaces all bookmarks in the index
func (idx *MemoryIndex) UpdateBookmarks(bookmarks []*domain.Bookmark) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear and rebuild
	idx.bookmarks = make(map[string]*domain.Bookmark, len(bookmarks))
	for _, bookmark := range bookmarks {
		idx.bookmarks[bookmark.ID] = bookmark
	}
	idx.lastBookmarkReload = time.Now()
}

// GetBookmark retrieves a bookmark by ID
func (idx *MemoryIndex) GetBookmark(id string) (*domain.Bookmark, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	bookmark, ok := idx.bookmarks[id]
	return bookmark, ok
}

// GetAllBookmarks returns all bookmarks
func (idx *MemoryIndex) GetAllBookmarks() []*domain.Bookmark {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	bookmarks := make([]*domain.Bookmark, 0, len(idx.bookmarks))
	for _, bookmark := range idx.bookmarks {
		bookmarks = append(bookmarks, bookmark)
	}
	return bookmarks
}

// AddBookmark adds or updates a single bookmark
func (idx *MemoryIndex) AddBookmark(bookmark *domain.Bookmark) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.bookmarks[bookmark.ID] = bookmark
}

// DeleteBookmark removes a bookmark from the index
func (idx *MemoryIndex) DeleteBookmark(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.bookmarks, id)
}

// BookmarkCount returns the number of bookmarks in the index
func (idx *MemoryIndex) BookmarkCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.bookmarks)
}

// GetLastBookmarkReload returns the timestamp of the last bookmarks reload
func (idx *MemoryIndex) GetLastBookmarkReload() time.Time {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.lastBookmarkReload
}
