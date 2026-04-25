use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

#[derive(Debug, Clone, PartialEq)]
pub enum CircuitState {
    Closed,
    Open,
    HalfOpen,
}

#[derive(Debug, thiserror::Error)]
pub enum CircuitBreakerError<E> {
    #[error("Circuit breaker is open")]
    Open,
    #[error("Wrapped error: {0}")]
    Inner(#[from] E),
}

pub struct CircuitBreaker {
    state: Arc<RwLock<CircuitState>>,
    failure_count: Arc<RwLock<u32>>,
    failure_threshold: u32,
    timeout: Duration,
    last_failure_time: Arc<RwLock<Option<Instant>>>,
}

impl CircuitBreaker {
    pub fn new(failure_threshold: u32, timeout: Duration) -> Self {
        Self {
            state: Arc::new(RwLock::new(CircuitState::Closed)),
            failure_count: Arc::new(RwLock::new(0)),
            failure_threshold,
            timeout,
            last_failure_time: Arc::new(RwLock::new(None)),
        }
    }

    pub async fn execute<F, T, E>(&self, operation: F) -> Result<T, CircuitBreakerError<E>>
    where
        F: std::future::Future<Output = Result<T, E>>,
    {
        // Check if circuit is open and should attempt to close
        self.check_state().await;

        let state = *self.state.read().await;

        if state == CircuitState::Open {
            return Err(CircuitBreakerError::Open);
        }

        // Execute the operation
        let result = operation.await;

        match result {
            Ok(value) => {
                self.on_success().await;
                Ok(value)
            }
            Err(error) => {
                self.on_failure().await;
                Err(CircuitBreakerError::Inner(error))
            }
        }
    }

    async fn check_state(&self) {
        let state = *self.state.read().await;
        let last_failure = *self.last_failure_time.read().await;

        if state == CircuitState::Open {
            if let Some(failure_time) = last_failure {
                if failure_time.elapsed() > self.timeout {
                    // Transition to half-open
                    *self.state.write().await = CircuitState::HalfOpen;
                    *self.failure_count.write().await = 0;
                }
            }
        }
    }

    async fn on_success(&self) {
        let state = *self.state.read().await;

        match state {
            CircuitState::HalfOpen => {
                // Reset to closed on success in half-open
                *self.state.write().await = CircuitState::Closed;
                *self.failure_count.write().await = 0;
            }
            CircuitState::Closed => {
                *self.failure_count.write().await = 0;
            }
            _ => {}
        }
    }

    async fn on_failure(&self) {
        let mut failure_count = self.failure_count.write().await;
        *failure_count += 1;

        if *failure_count >= self.failure_threshold {
            *self.state.write().await = CircuitState::Open;
            *self.last_failure_time.write().await = Some(Instant::now());
        }
    }

    pub async fn state(&self) -> CircuitState {
        *self.state.read().await
    }

    pub async fn reset(&self) {
        *self.state.write().await = CircuitState::Closed;
        *self.failure_count.write().await = 0;
        *self.last_failure_time.write().await = None;
    }
}

impl Default for CircuitBreaker {
    fn default() -> Self {
        Self::new(5, Duration::from_secs(60))
    }
}
