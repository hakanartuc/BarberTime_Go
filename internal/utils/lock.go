package utils

import (
	"barbertime_go/pkg/redis"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// DistributedLock distributed lock wrapper
type DistributedLock struct {
	key   string
	value string
	ttl   time.Duration
}

// NewDistributedLock yeni bir distributed lock oluşturur
func NewDistributedLock(key string, ttl time.Duration) *DistributedLock {
	// Unique lock value oluştur (16 byte random)
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	lockValue := hex.EncodeToString(randomBytes)

	return &DistributedLock{
		key:   key,
		value: lockValue,
		ttl:   ttl,
	}
}

// Acquire lock'u alır, başarılı olursa true döner
func (dl *DistributedLock) Acquire() (bool, error) {
	return redis.AcquireLock(dl.key, dl.value, dl.ttl)
}

// Release lock'u serbest bırakır
func (dl *DistributedLock) Release() error {
	return redis.ReleaseLock(dl.key, dl.value)
}

// Extend lock'un süresini uzatır
func (dl *DistributedLock) Extend(newTTL time.Duration) (bool, error) {
	return redis.ExtendLock(dl.key, dl.value, newTTL)
}

// WithLock lock alır, fonksiyonu çalıştırır ve lock'u serbest bırakır
func WithLock(lockKey string, ttl time.Duration, fn func() error) error {
	lock := NewDistributedLock(lockKey, ttl)

	// Lock al
	acquired, err := lock.Acquire()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !acquired {
		return fmt.Errorf("could not acquire lock, resource is locked")
	}

	// İşlem bittiğinde lock'u serbest bırak
	defer func() {
		if releaseErr := lock.Release(); releaseErr != nil {
			// Log error ama fonksiyonun hatasını değiştirme
			fmt.Printf("Warning: Failed to release lock: %v\n", releaseErr)
		}
	}()

	// Fonksiyonu çalıştır
	return fn()
}
