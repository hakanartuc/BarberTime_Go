package redis

import (
	"barbertime_go/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

func Connect() error {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
		Password: config.AppConfig.RedisPassword,
		DB:       config.AppConfig.RedisDB,
	})

	// Bağlantıyı test et
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

func Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return Client.Set(ctx, key, data, expiration).Err()
}

func Get(key string, dest interface{}) error {
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

func Delete(key string) error {
	return Client.Del(ctx, key).Err()
}

func Exists(key string) (bool, error) {
	count, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func SetString(key string, value string, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

func GetString(key string) (string, error) {
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}
	return val, nil
}

func Increment(key string) (int64, error) {
	return Client.Incr(ctx, key).Result()
}

func Expire(key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

// Distributed Lock fonksiyonları
type Lock struct {
	key   string
	value string
}

// AcquireLock distributed lock alır (SET NX kullanarak)
// lockKey: Lock için kullanılacak key
// lockValue: Lock sahibini belirten unique değer (genellikle UUID veya process ID)
// ttl: Lock'un otomatik olarak serbest bırakılacağı süre
func AcquireLock(lockKey string, lockValue string, ttl time.Duration) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	// SET key value NX EX seconds
	// NX: Sadece key yoksa set et
	// EX: Expiration time (seconds)
	result, err := Client.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return result, nil
}

// ReleaseLock distributed lock'u serbest bırakır
// Lua script kullanarak sadece lock sahibi lock'u serbest bırakabilir
func ReleaseLock(lockKey string, lockValue string) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// Lua script: Sadece lock value eşleşirse sil
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	_, err := Client.Eval(ctx, script, []string{lockKey}, lockValue).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

// ExtendLock lock'un süresini uzatır
func ExtendLock(lockKey string, lockValue string, ttl time.Duration) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	// Lua script: Sadece lock value eşleşirse expire'ı uzat
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := Client.Eval(ctx, script, []string{lockKey}, lockValue, int(ttl.Seconds())).Result()
	if err != nil {
		return false, fmt.Errorf("failed to extend lock: %w", err)
	}

	return result.(int64) == 1, nil
}
