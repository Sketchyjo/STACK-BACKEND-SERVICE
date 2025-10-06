package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	walletprovisioning "github.com/stack-service/stack_service/internal/workers/wallet_provisioning"
	"go.uber.org/zap"
)

// WorkerHandlers contains the worker management HTTP handlers
type WorkerHandlers struct {
	scheduler interface{} // Will be *walletprovisioning.Scheduler at runtime
	logger    *zap.Logger
}

// NewWorkerHandlers creates a new instance of worker handlers
func NewWorkerHandlers(scheduler interface{}, logger *zap.Logger) *WorkerHandlers {
	return &WorkerHandlers{
		scheduler: scheduler,
		logger:    logger,
	}
}

// GetWorkerStatus handles GET /admin/workers/status
// @Summary Get worker status
// @Description Returns the current status of the wallet provisioning worker and scheduler
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{} "Worker status"
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/workers/status [get]
func (h *WorkerHandlers) GetWorkerStatus(c *gin.Context) {
	h.logger.Debug("Getting worker status")

	// Cast to scheduler type
	scheduler, ok := h.scheduler.(*walletprovisioning.Scheduler)
	if !ok || scheduler == nil {
		h.logger.Error("Scheduler not available or wrong type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SCHEDULER_UNAVAILABLE",
			"message": "Worker scheduler is not available",
		})
		return
	}

	// Get scheduler status
	status := scheduler.GetStatus()

	h.logger.Debug("Retrieved worker status", zap.Bool("is_running", status.IsRunning))

	c.JSON(http.StatusOK, gin.H{
		"worker": gin.H{
			"type":   "wallet_provisioning",
			"status": "operational",
		},
		"scheduler": gin.H{
			"is_running":      status.IsRunning,
			"poll_interval":   status.PollInterval.String(),
			"max_concurrency": status.MaxConcurrency,
			"active_jobs":     status.ActiveJobs,
		},
		"metrics": gin.H{
			"total_jobs_processed": status.WorkerMetrics.TotalJobsProcessed,
			"successful_jobs":      status.WorkerMetrics.SuccessfulJobs,
			"failed_jobs":          status.WorkerMetrics.FailedJobs,
			"total_retries":        status.WorkerMetrics.TotalRetries,
			"average_duration_ms":  status.WorkerMetrics.AverageDuration.Milliseconds(),
			"last_processed_at":    status.WorkerMetrics.LastProcessedAt,
			"errors_by_type":       status.WorkerMetrics.ErrorsByType,
		},
	})
}

// GetWorkerMetrics handles GET /admin/workers/metrics
// @Summary Get worker metrics
// @Description Returns detailed metrics for the wallet provisioning worker
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{} "Worker metrics"
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/workers/metrics [get]
func (h *WorkerHandlers) GetWorkerMetrics(c *gin.Context) {
	h.logger.Debug("Getting worker metrics")

	// Cast to scheduler type
	scheduler, ok := h.scheduler.(*walletprovisioning.Scheduler)
	if !ok || scheduler == nil {
		h.logger.Error("Scheduler not available or wrong type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SCHEDULER_UNAVAILABLE",
			"message": "Worker scheduler is not available",
		})
		return
	}

	// Get scheduler status
	status := scheduler.GetStatus()
	metrics := status.WorkerMetrics

	// Calculate success rate
	var successRate float64
	if metrics.TotalJobsProcessed > 0 {
		successRate = float64(metrics.SuccessfulJobs) / float64(metrics.TotalJobsProcessed) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total_jobs_processed": metrics.TotalJobsProcessed,
		"successful_jobs":      metrics.SuccessfulJobs,
		"failed_jobs":          metrics.FailedJobs,
		"success_rate":         successRate,
		"total_retries":        metrics.TotalRetries,
		"average_duration": gin.H{
			"milliseconds": metrics.AverageDuration.Milliseconds(),
			"seconds":      metrics.AverageDuration.Seconds(),
		},
		"last_processed_at": metrics.LastProcessedAt,
		"errors_by_type":    metrics.ErrorsByType,
		"active_jobs":       status.ActiveJobs,
	})
}

// TriggerJobProcessing handles POST /admin/workers/trigger
// @Summary Trigger job processing
// @Description Manually triggers processing of a specific wallet provisioning job
// @Tags admin
// @Accept json
// @Produce json
// @Param request body TriggerJobRequest true "Job ID to process"
// @Success 202 {object} map[string]interface{} "Job processing triggered"
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/workers/trigger [post]
func (h *WorkerHandlers) TriggerJobProcessing(c *gin.Context) {
	h.logger.Info("Manual job processing triggered",
		zap.String("request_id", getRequestID(c)),
		zap.String("ip", c.ClientIP()))

	var req TriggerJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid trigger request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	// Validate job ID
	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		h.logger.Warn("Invalid job ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_JOB_ID",
			"message": "Invalid job ID format",
			"details": gin.H{"job_id": req.JobID},
		})
		return
	}

	h.logger.Info("Triggering job processing",
		zap.String("job_id", jobID.String()))

	// Note: For immediate processing, you would need to expose a method on the worker
	// For now, return accepted status indicating the job will be picked up by scheduler
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Job will be processed by the scheduler",
		"job_id":  jobID.String(),
		"note":    "The job will be picked up in the next scheduler poll cycle",
	})
}

// GetWorkerHealth handles GET /admin/workers/health
// @Summary Worker health check
// @Description Returns health status of the wallet provisioning worker
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status"
// @Failure 503 {object} entities.ErrorResponse
// @Router /api/v1/admin/workers/health [get]
func (h *WorkerHandlers) GetWorkerHealth(c *gin.Context) {
	h.logger.Debug("Worker health check requested")

	// Cast to scheduler type
	scheduler, ok := h.scheduler.(*walletprovisioning.Scheduler)
	if !ok || scheduler == nil {
		h.logger.Error("Scheduler not available")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   "SCHEDULER_UNAVAILABLE",
			"message": "Worker scheduler is not available",
		})
		return
	}

	// Check if scheduler is running
	isRunning := scheduler.IsRunning()

	if !isRunning {
		h.logger.Warn("Scheduler is not running")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"reason":  "scheduler_not_running",
			"message": "Wallet provisioning scheduler is not running",
		})
		return
	}

	// Get metrics to check if worker is processing jobs
	status := scheduler.GetStatus()

	h.logger.Debug("Worker health check passed")

	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"scheduler":   "running",
		"active_jobs": status.ActiveJobs,
		"metrics": gin.H{
			"total_processed": status.WorkerMetrics.TotalJobsProcessed,
			"last_activity":   status.WorkerMetrics.LastProcessedAt,
		},
	})
}

// RestartScheduler handles POST /admin/workers/restart
// @Summary Restart worker scheduler
// @Description Stops and restarts the wallet provisioning scheduler
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{} "Scheduler restarted"
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/workers/restart [post]
func (h *WorkerHandlers) RestartScheduler(c *gin.Context) {
	h.logger.Info("Scheduler restart requested",
		zap.String("request_id", getRequestID(c)),
		zap.String("ip", c.ClientIP()))

	// Cast to scheduler type
	scheduler, ok := h.scheduler.(*walletprovisioning.Scheduler)
	if !ok || scheduler == nil {
		h.logger.Error("Scheduler not available")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SCHEDULER_UNAVAILABLE",
			"message": "Worker scheduler is not available",
		})
		return
	}

	// Stop scheduler
	if err := scheduler.Stop(); err != nil {
		h.logger.Error("Failed to stop scheduler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "STOP_FAILED",
			"message": "Failed to stop scheduler",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		h.logger.Error("Failed to start scheduler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "START_FAILED",
			"message": "Failed to start scheduler",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	h.logger.Info("Scheduler restarted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduler restarted successfully",
		"status":  "running",
	})
}

// Request models

// TriggerJobRequest represents a request to trigger job processing
type TriggerJobRequest struct {
	JobID string `json:"job_id" validate:"required,uuid"`
}
