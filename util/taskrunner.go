package util

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus int

const (
	TaskPending   TaskStatus = iota // Not yet started
	TaskRunning                     // Currently executing
	TaskCompleted                   // Finished successfully
	TaskFailed                      // Finished with error
	TaskCancelled                   // Cancelled by user
)

// String returns a human-readable status name
func (s TaskStatus) String() string {
	switch s {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskCompleted:
		return "completed"
	case TaskFailed:
		return "failed"
	case TaskCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Task represents a background operation
type Task struct {
	ID          string
	Name        string
	Status      TaskStatus
	Progress    float64 // 0.0 to 1.0 (-1 for indeterminate)
	Message     string  // Current status message
	Error       error   // Error if failed
	StartedAt   time.Time
	CompletedAt time.Time
	Result      any // Result data on completion

	cancel context.CancelFunc
	mu     sync.RWMutex
}

// IsRunning returns true if the task is currently executing
func (t *Task) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status == TaskRunning
}

// IsComplete returns true if the task has finished (success, failure, or cancelled)
func (t *Task) IsComplete() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status == TaskCompleted || t.Status == TaskFailed || t.Status == TaskCancelled
}

// Duration returns how long the task has been running (or total duration if complete)
func (t *Task) Duration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.StartedAt.IsZero() {
		return 0
	}

	if !t.CompletedAt.IsZero() {
		return t.CompletedAt.Sub(t.StartedAt)
	}

	return time.Since(t.StartedAt)
}

// TaskFunc is the function signature for task execution
// - ctx: for cancellation
// - progress: callback to report progress (0.0-1.0) and message
type TaskFunc func(ctx context.Context, progress func(pct float64, msg string)) (any, error)

// TaskRunner manages background tasks
type TaskRunner struct {
	tasks       map[string]*Task
	mu          sync.RWMutex
	maxParallel int // Max concurrent tasks (0 = unlimited)
	running     int // Current running count
	queue       []*queuedTask

	// Callbacks
	onStart    func(task *Task)
	onProgress func(task *Task)
	onComplete func(task *Task)
	onError    func(task *Task)
}

type queuedTask struct {
	task *Task
	fn   TaskFunc
}

// NewTaskRunner creates a new task runner
func NewTaskRunner() *TaskRunner {
	return &TaskRunner{
		tasks: make(map[string]*Task),
	}
}

// SetMaxParallel limits concurrent task execution
// 0 = unlimited (default)
func (r *TaskRunner) SetMaxParallel(max int) *TaskRunner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.maxParallel = max
	return r
}

// SetOnStart is called when a task begins execution
func (r *TaskRunner) SetOnStart(fn func(task *Task)) *TaskRunner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onStart = fn
	return r
}

// SetOnProgress is called when task reports progress
func (r *TaskRunner) SetOnProgress(fn func(task *Task)) *TaskRunner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onProgress = fn
	return r
}

// SetOnComplete is called when task finishes successfully
func (r *TaskRunner) SetOnComplete(fn func(task *Task)) *TaskRunner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onComplete = fn
	return r
}

// SetOnError is called when task fails
func (r *TaskRunner) SetOnError(fn func(task *Task)) *TaskRunner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onError = fn
	return r
}

// Run starts a new background task
// Returns the task immediately; execution is async
func (r *TaskRunner) Run(name string, fn TaskFunc) *Task {
	return r.RunWithID(generateTaskID(), name, fn)
}

// RunWithID starts a task with a specific ID (for tracking)
func (r *TaskRunner) RunWithID(id, name string, fn TaskFunc) *Task {
	ctx, cancel := context.WithCancel(context.Background())

	task := &Task{
		ID:       id,
		Name:     name,
		Status:   TaskPending,
		Progress: -1, // Indeterminate by default
		cancel:   cancel,
	}

	r.mu.Lock()
	r.tasks[id] = task

	// Check if we can start immediately or need to queue
	if r.maxParallel > 0 && r.running >= r.maxParallel {
		r.queue = append(r.queue, &queuedTask{task: task, fn: fn})
		r.mu.Unlock()
		return task
	}

	r.running++
	r.mu.Unlock()

	// Start the task
	go r.executeTask(ctx, task, fn)

	return task
}

func (r *TaskRunner) executeTask(ctx context.Context, task *Task, fn TaskFunc) {
	// Mark as running
	task.mu.Lock()
	task.Status = TaskRunning
	task.StartedAt = time.Now()
	task.mu.Unlock()

	// Notify start
	r.mu.RLock()
	onStart := r.onStart
	r.mu.RUnlock()
	if onStart != nil {
		onStart(task)
	}

	// Progress callback
	progress := func(pct float64, msg string) {
		task.mu.Lock()
		task.Progress = pct
		task.Message = msg
		task.mu.Unlock()

		r.mu.RLock()
		onProgress := r.onProgress
		r.mu.RUnlock()
		if onProgress != nil {
			onProgress(task)
		}
	}

	// Execute the task
	result, err := fn(ctx, progress)

	// Update task state
	task.mu.Lock()
	task.CompletedAt = time.Now()
	task.Result = result

	if ctx.Err() != nil {
		task.Status = TaskCancelled
		task.Error = ctx.Err()
	} else if err != nil {
		task.Status = TaskFailed
		task.Error = err
	} else {
		task.Status = TaskCompleted
		task.Progress = 1.0
	}
	status := task.Status
	task.mu.Unlock()

	// Notify completion
	r.mu.RLock()
	onComplete := r.onComplete
	onError := r.onError
	r.mu.RUnlock()

	if status == TaskCompleted && onComplete != nil {
		onComplete(task)
	} else if (status == TaskFailed || status == TaskCancelled) && onError != nil {
		onError(task)
	}

	// Start next queued task
	r.mu.Lock()
	r.running--
	if len(r.queue) > 0 {
		next := r.queue[0]
		r.queue = r.queue[1:]
		r.running++
		go r.executeTask(ctx, next.task, next.fn)
	}
	r.mu.Unlock()
}

// Cancel requests cancellation of a running task
func (r *TaskRunner) Cancel(id string) error {
	r.mu.RLock()
	task, exists := r.tasks[id]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	task.mu.RLock()
	if task.Status != TaskRunning && task.Status != TaskPending {
		task.mu.RUnlock()
		return fmt.Errorf("task is not running: %s", id)
	}
	task.mu.RUnlock()

	if task.cancel != nil {
		task.cancel()
	}

	return nil
}

// CancelAll cancels all running tasks
func (r *TaskRunner) CancelAll() {
	r.mu.RLock()
	tasks := make([]*Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	r.mu.RUnlock()

	for _, task := range tasks {
		task.mu.RLock()
		isRunning := task.Status == TaskRunning || task.Status == TaskPending
		task.mu.RUnlock()

		if isRunning && task.cancel != nil {
			task.cancel()
		}
	}
}

// Get returns a task by ID
func (r *TaskRunner) Get(id string) *Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tasks[id]
}

// GetAll returns all tasks
func (r *TaskRunner) GetAll() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetRunning returns all currently running tasks
func (r *TaskRunner) GetRunning() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []*Task
	for _, task := range r.tasks {
		task.mu.RLock()
		if task.Status == TaskRunning {
			tasks = append(tasks, task)
		}
		task.mu.RUnlock()
	}
	return tasks
}

// IsRunning checks if a specific task is running
func (r *TaskRunner) IsRunning(id string) bool {
	r.mu.RLock()
	task, exists := r.tasks[id]
	r.mu.RUnlock()

	if !exists {
		return false
	}

	return task.IsRunning()
}

// Wait blocks until a task completes
func (r *TaskRunner) Wait(id string) *Task {
	r.mu.RLock()
	task, exists := r.tasks[id]
	r.mu.RUnlock()

	if !exists {
		return nil
	}

	// Poll until complete
	for !task.IsComplete() {
		time.Sleep(10 * time.Millisecond)
	}

	return task
}

// WaitAll blocks until all tasks complete
func (r *TaskRunner) WaitAll() {
	for {
		r.mu.RLock()
		allComplete := true
		for _, task := range r.tasks {
			if !task.IsComplete() {
				allComplete = false
				break
			}
		}
		r.mu.RUnlock()

		if allComplete {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// Remove removes a completed/failed/cancelled task from tracking
func (r *TaskRunner) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		return
	}

	task.mu.RLock()
	isComplete := task.Status == TaskCompleted || task.Status == TaskFailed || task.Status == TaskCancelled
	task.mu.RUnlock()

	if isComplete {
		delete(r.tasks, id)
	}
}

// Clear removes all completed tasks
func (r *TaskRunner) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, task := range r.tasks {
		task.mu.RLock()
		isComplete := task.Status == TaskCompleted || task.Status == TaskFailed || task.Status == TaskCancelled
		task.mu.RUnlock()

		if isComplete {
			delete(r.tasks, id)
		}
	}
}

// Count returns the number of tasks in each status
func (r *TaskRunner) Count() map[TaskStatus]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := make(map[TaskStatus]int)
	for _, task := range r.tasks {
		task.mu.RLock()
		counts[task.Status]++
		task.mu.RUnlock()
	}
	return counts
}

// generateTaskID generates a unique task ID
func generateTaskID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("task_%d_%s", time.Now().UnixNano(), hex.EncodeToString(bytes))
}
