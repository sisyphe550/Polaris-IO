package svc

import (
	"sync"
	"sync/atomic"
	"time"

	"shared-board/backend/app/user/model"
)

// MemUserStore: 当 DB 未配置时用于本地调试的内存用户仓库。
// 注意：进程重启后数据会丢失，仅用于开发/联调。
type MemUserStore struct {
	mu         sync.RWMutex
	nextUserID uint64
	byUsername map[string]*model.Users
	byID       map[uint64]*model.Users
}

func NewMemUserStore() *MemUserStore {
	return &MemUserStore{
		nextUserID: 0,
		byUsername: make(map[string]*model.Users),
		byID:       make(map[uint64]*model.Users),
	}
}

func (s *MemUserStore) FindOneByUsername(username string) (*model.Users, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byUsername[username]
	return u, ok
}

func (s *MemUserStore) FindOne(id uint64) (*model.Users, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	return u, ok
}

func (s *MemUserStore) Insert(username, passwordHash string) uint64 {
	now := time.Now()
	id := atomic.AddUint64(&s.nextUserID, 1)

	u := &model.Users{
		Id:         id,
		Username:   username,
		Password:   passwordHash,
		Avatar:     "",
		Info:       "",
		CreateTime: now,
		UpdateTime: now,
		DeleteTime: time.Unix(0, 0),
		DelState:   0,
		Version:    1,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.byUsername[username] = u
	s.byID[id] = u
	return id
}
