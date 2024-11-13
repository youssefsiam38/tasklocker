package tasklocker

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// AcquireLock tries to acquire a lock for concurrent tasks using Redis.
// It returns true if the lock is successfully acquired, otherwise false.
// Parameters:
// - ctx: The context for the Redis operations.
// - client: The Redis client instance.
// - key: The key used to track the concurrent tasks in Redis.
// - allowedConcurrentTasks: The maximum number of concurrent tasks allowed.
// - timeout: The duration after which the lock should be automatically released.
func AcquireLock(ctx context.Context, client *redis.Client, key string, allowedConcurrentTasks int, timeout time.Duration) (bool, error) {
	luaScript := `
		local currentVal = redis.call("GET", KEYS[1])
		if not currentVal then
			currentVal = 0
		end
		if tonumber(currentVal) < tonumber(ARGV[1]) then
			redis.call("INCR", KEYS[1])
			redis.call("PEXPIRE", KEYS[1], ARGV[2])
			return true
		else
			return false
		end
	`
	// Execute the Lua script to ensure atomicity
	result, err := client.Eval(ctx, luaScript, []string{key}, allowedConcurrentTasks, int(timeout.Milliseconds())).Bool()
	if err != nil {
		return false, err
	}
	return result, nil
}

// ReleaseLock releases the lock for concurrent tasks by decrementing the counter in Redis.
// Parameters:
// - ctx: The context for the Redis operations.
// - client: The Redis client instance.
// - key: The key used to track the concurrent tasks in Redis.
func ReleaseLock(ctx context.Context, client *redis.Client, key string) error {
	luaScript := `
		local currentVal = redis.call("GET", KEYS[1])
		if currentVal and tonumber(currentVal) > 0 then
			redis.call("DECR", KEYS[1])
		end
	`
	// Execute the Lua script to ensure atomicity
	return client.Eval(ctx, luaScript, []string{key}).Err()
}
