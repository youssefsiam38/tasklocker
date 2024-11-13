package tasklocker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// AcquireLock tries to acquire a lock for concurrent tasks using Redis.
// It returns a boolean indicating whether the lock is acquired, a boolean indicating whether the key exists,
// and an error if something goes wrong.
// Parameters:
// - ctx: The context for the Redis operations.
// - client: The Redis client instance.
// - prefix: The prefix for the task key.
// - postfix: The unique identifier for the task (e.g., task id).
// - allowedConcurrentTasks: The maximum number of concurrent tasks allowed.
// - timeout: The duration after which the lock should be automatically released.
func AcquireLock(ctx context.Context, client *redis.Client, prefix, postfix string, allowedConcurrentTasks int, timeout time.Duration) (bool, bool, error) {
	// Create the task-specific key using the prefix and postfix (e.g., google_places_brands_processor:1)
	taskKey := fmt.Sprintf("%s:%s", prefix, postfix)

	// Check if the specific key already exists
	exists, err := client.Exists(ctx, taskKey).Result()
	if err != nil {
		return false, false, fmt.Errorf("failed to check if key exists: %v", err)
	}
	if exists > 0 {
		// The key exists, return true for "exist"
		return false, true, nil
	}

	// Count how many tasks are currently active (matching the prefix)
	keys, err := client.Keys(ctx, fmt.Sprintf("%s:*", prefix)).Result()
	if err != nil {
		return false, false, fmt.Errorf("failed to get keys with prefix: %v", err)
	}

	// If the number of active tasks exceeds the allowedConcurrentTasks, do not acquire the lock
	if len(keys) >= allowedConcurrentTasks {
		return false, false, nil // Lock cannot be acquired
	}

	// Try to acquire the lock for the task by setting a key with an expiration time
	err = client.SetEx(ctx, taskKey, 1, timeout).Err()
	if err != nil {
		return false, false, fmt.Errorf("failed to set key: %v", err)
	}

	return true, false, nil // Lock acquired successfully
}

// ReleaseLock releases the lock for concurrent tasks by decrementing the counter in Redis.
// Parameters:
// - ctx: The context for the Redis operations.
// - client: The Redis client instance.
// - prefix: The prefix for the task key.
// - postfix: The unique identifier for the task (e.g., task id).
func ReleaseLock(ctx context.Context, client *redis.Client, prefix, postfix string) error {
	// Construct the task key using the prefix and postfix (e.g., google_places_brands_processor:1)
	taskKey := fmt.Sprintf("%s:%s", prefix, postfix)

	// Delete the task-specific key to release the lock
	return client.Del(ctx, taskKey).Err()
}
