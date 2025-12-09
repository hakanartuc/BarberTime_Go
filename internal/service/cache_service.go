package service

import (
	"barbertime_go/pkg/redis"
	"time"
)

type CacheService interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}

type cacheService struct{}

func NewCacheService() CacheService {
	return &cacheService{}
}

func (s *cacheService) Get(key string, dest interface{}) error {
	if redis.Client == nil {
		return redis.Get(key, dest)
	}
	return redis.Get(key, dest)
}

func (s *cacheService) Set(key string, value interface{}, expiration time.Duration) error {
	if redis.Client == nil {
		return nil // Redis yoksa sessizce geç
	}
	return redis.Set(key, value, expiration)
}

func (s *cacheService) Delete(key string) error {
	if redis.Client == nil {
		return nil
	}
	return redis.Delete(key)
}
