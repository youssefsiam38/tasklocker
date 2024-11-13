# TaskLocker

`TaskLocker` is a Go package that provides a simple and effective way to manage concurrent task execution using Redis. It ensures that no more than a specified number of tasks run concurrently by leveraging Redis's atomic operations and Lua scripting.

## Features

- **Concurrency Control**: Limit the number of concurrent tasks using Redis.
- **Automatic Timeout**: Locks automatically expire after a specified duration, ensuring tasks don't hold the lock indefinitely.
- **Atomic Operations**: Uses Redis commands to ensure thread safety.

## Installation

To install the package, use `go get`:

```bash
go get github.com/youssefsiam38/tasklocker
```

## Usage

Here's a basic example demonstrating how to use the `TaskLocker` package:

### Import the package

```go
import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/youssefsiam38/tasklocker"
)
```

### Example Code

```go
func main() {
    ctx := context.Background()
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379", // Redis server address
    })
    prefix := "google_places_brands_processor"
    allowedConcurrentTasks := 3
    timeout := 5 * time.Second

    var wg sync.WaitGroup

    // Function to simulate task execution
    runTask := func(taskID int) {
        defer wg.Done()
        postfix := fmt.Sprintf("%d", taskID)
        success, err := tasklocker.AcquireLock(ctx, client, prefix, postfix, allowedConcurrentTasks, timeout)
        if err != nil {
            log.Fatalf("failed to acquire lock for task %d: %v", taskID, err)
        }
        if success {
            fmt.Printf("Task %d acquired the lock\n", taskID)
            // Simulate task duration
            time.Sleep(timeout / 2) // Run task for half of the timeout duration
            if err := tasklocker.ReleaseLock(ctx, client, prefix, postfix); err != nil {
                log.Fatalf("failed to release lock for task %d: %v", taskID, err)
            }
            fmt.Printf("Task %d released the lock\n", taskID)
        } else {
            fmt.Printf("Task %d could not acquire the lock\n", taskID)
        }
    }

    // First round of tasks
    for i := 0; i < 6; i++ {
        wg.Add(1)
        go runTask(i)
    }

    // Wait for the first round of tasks to finish
    wg.Wait()

    fmt.Println("Waiting for the timeout period to elapse...")

    // Wait for the timeout period to elapse
    time.Sleep(timeout + 1*time.Second)

    // Second round of tasks to ensure tasks can be acquired again after timeout
    for i := 6; i < 12; i++ {
        wg.Add(1)
        go runTask(i)
    }

    // Wait for the second round of tasks to complete
    wg.Wait()
}
```

## Functions

### `AcquireLock`

```go
func AcquireLock(ctx context.Context, client *redis.Client, prefix, postfix string, allowedConcurrentTasks int, timeout time.Duration) (bool, error)
```

Attempts to acquire a lock for concurrent tasks using Redis. Returns `true` if the lock is successfully acquired, otherwise `false`.

- **Parameters**:
  - `ctx`: The context for the Redis operations.
  - `client`: The Redis client instance.
  - `prefix`: The prefix for the task key.
  - `postfix`: The unique identifier for the task (e.g., task id).
  - `allowedConcurrentTasks`: The maximum number of concurrent tasks allowed.
  - `timeout`: The duration after which the lock should be automatically released.

### `ReleaseLock`

```go
func ReleaseLock(ctx context.Context, client *redis.Client, prefix, postfix string) error
```

Releases the lock for concurrent tasks by deleting the task-specific key in Redis.

- **Parameters**:
  - `ctx`: The context for the Redis operations.
  - `client`: The Redis client instance.
  - `prefix`: The prefix for the task key.
  - `postfix`: The unique identifier for the task (e.g., task id).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with your changes.

## Author

Youssef Siam - [GitHub](https://github.com/youssefsiam38)
