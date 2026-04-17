package services

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
)

// ChaosService implements chaos engineering experiments
type ChaosService struct {
	db     *database.Database
	active map[string]*Experiment
	mu     sync.RWMutex
}

type Experiment struct {
	ID       string
	Name     string
	Target   string
	Type     ExperimentType
	Config   ExperimentConfig
	Status   ExperimentStatus
	Started  time.Time
	Finished *time.Time
	Error    string
}

type ExperimentType string

const (
	ExperimentTypeLatency   ExperimentType = "latency"
	ExperimentTypeKill      ExperimentType = "kill"
	ExperimentTypeCrash     ExperimentType = "crash"
	ExperimentTypePartition ExperimentType = "partition"
	ExperimentTypeFlush     ExperimentType = "flush"
	ExperimentTypeLoad      ExperimentType = "load"
)

type ExperimentStatus string

const (
	StatusIdle      ExperimentStatus = "idle"
	StatusRunning   ExperimentStatus = "running"
	StatusCompleted ExperimentStatus = "completed"
	StatusFailed    ExperimentStatus = "failed"
)

type ExperimentConfig struct {
	Duration    time.Duration
	Probability float64
	Amount      int
	Target      string
}

type ChaosMiddleware struct {
	service *ChaosService
}

// NewChaosService creates a new chaos engineering service
func NewChaosService(db *database.Database) *ChaosService {
	return &ChaosService{
		db:     db,
		active: make(map[string]*Experiment),
	}
}

// RunExperiment starts a chaos experiment
func (cs *ChaosService) RunExperiment(ctx context.Context, expType ExperimentType, config ExperimentConfig) (*Experiment, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	id := fmt.Sprintf("exp_%d", time.Now().Unix())
	exp := &Experiment{
		ID:      id,
		Name:    fmt.Sprintf("%s Experiment", expType),
		Target:  config.Target,
		Type:    expType,
		Config:  config,
		Status:  StatusRunning,
		Started: time.Now(),
	}

	cs.active[id] = exp

	// Run experiment in background
	go cs.executeExperiment(ctx, exp)

	return exp, nil
}

func (cs *ChaosService) executeExperiment(ctx context.Context, exp *Experiment) {
	defer func() {
		cs.mu.Lock()
		finished := time.Now()
		exp.Finished = &finished
		if exp.Status == StatusRunning {
			exp.Status = StatusCompleted
		}
		delete(cs.active, exp.ID)
		cs.mu.Unlock()
	}()

	switch exp.Type {
	case ExperimentTypeLatency:
		cs.injectLatency(ctx, exp)
	case ExperimentTypeKill:
		cs.killConnections(ctx, exp)
	case ExperimentTypeCrash:
		cs.crashService(ctx, exp)
	case ExperimentTypePartition:
		cs.networkPartition(ctx, exp)
	case ExperimentTypeFlush:
		cs.flushCache(ctx, exp)
	case ExperimentTypeLoad:
		cs.simulateLoad(ctx, exp)
	default:
		cs.mu.Lock()
		exp.Status = StatusFailed
		exp.Error = "unknown experiment type"
		cs.mu.Unlock()
	}
}

func (cs *ChaosService) injectLatency(_ context.Context, exp *Experiment) {
	// Simulate latency injection
	delay := time.Duration(rand.Intn(1000)+100) * time.Millisecond
	time.Sleep(delay)

	cs.mu.Lock()
	exp.Status = StatusCompleted
	cs.mu.Unlock()
}

func (cs *ChaosService) killConnections(_ context.Context, exp *Experiment) {
	// Simulate database connection killing
	time.Sleep(exp.Config.Duration)

	cs.mu.Lock()
	exp.Status = StatusCompleted
	cs.mu.Unlock()
}

func (cs *ChaosService) crashService(_ context.Context, exp *Experiment) {
	// Simulate service crash
	time.Sleep(exp.Config.Duration)

	cs.mu.Lock()
	if rand.Float64() > 0.2 {
		exp.Status = StatusCompleted
	} else {
		exp.Status = StatusFailed
		exp.Error = "service crashed irrecoverably"
	}
	cs.mu.Unlock()
}

func (cs *ChaosService) networkPartition(_ context.Context, exp *Experiment) {
	// Simulate network partition
	time.Sleep(exp.Config.Duration)

	cs.mu.Lock()
	exp.Status = StatusCompleted
	cs.mu.Unlock()
}

func (cs *ChaosService) flushCache(_ context.Context, exp *Experiment) {
	// Simulate cache flush
	time.Sleep(exp.Config.Duration)

	cs.mu.Lock()
	exp.Status = StatusCompleted
	cs.mu.Unlock()
}

func (cs *ChaosService) simulateLoad(ctx context.Context, exp *Experiment) {
	// Simulate high load
	start := time.Now()
	for time.Since(start) < exp.Config.Duration {
		select {
		case <-ctx.Done():
			return
		default:
			// Simulate work
			time.Sleep(time.Millisecond * 10)
		}
	}

	cs.mu.Lock()
	exp.Status = StatusCompleted
	cs.mu.Unlock()
}

// GetActiveExperiments returns all currently running experiments
func (cs *ChaosService) GetActiveExperiments() []*Experiment {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	experiments := make([]*Experiment, 0, len(cs.active))
	for _, exp := range cs.active {
		experiments = append(experiments, exp)
	}
	return experiments
}

// GetExperimentStatus returns the status of a specific experiment
func (cs *ChaosService) GetExperimentStatus(id string) (*Experiment, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	exp, exists := cs.active[id]
	return exp, exists
}

// Middleware function to inject chaos into HTTP requests
func (cm *ChaosMiddleware) InjectChaos(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cm.service.mu.RLock()
		for _, exp := range cm.service.active {
			if exp.Type == ExperimentTypeLatency && exp.Status == StatusRunning {
				// Inject random latency
				if rand.Float64() < exp.Config.Probability {
					delay := time.Duration(rand.Intn(int(exp.Config.Duration.Milliseconds()))) * time.Millisecond
					time.Sleep(delay)
				}
			}
		}
		cm.service.mu.RUnlock()

		next.ServeHTTP(w, r)
	})
}

// GetExperimentHistory returns a summary of past experiments (placeholder)
func (cs *ChaosService) GetExperimentHistory() ([]*Experiment, error) {
	// In a real implementation, this would query the database for experiment history
	return []*Experiment{}, nil
}

// StopExperiment stops a running experiment
func (cs *ChaosService) StopExperiment(id string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if exp, exists := cs.active[id]; exists {
		if exp.Status == StatusRunning {
			exp.Status = StatusCompleted
			finished := time.Now()
			exp.Finished = &finished
			delete(cs.active, id)
			return nil
		}
	}
	return fmt.Errorf("experiment %s not found or not running", id)
}
