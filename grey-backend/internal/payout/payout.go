package payout

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/ledger"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/models"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	bullmq "go.codycody31.dev/gobullmq"
)

const PayoutQueueName = "kite:payouts"
type Service struct {
	repo          *repository.Repository
	ledgerService *ledger.Service
	logger        *slog.Logger
	queue         *bullmq.Queue[PayoutJobData]
}
type PayoutJobData struct {
	PayoutID string `json:"payout_id"`
}

func NewService(repo *repository.Repository, ledgerService *ledger.Service, logger *slog.Logger, queueClient redis.Cmdable) (*Service, error) {
	queue, err := bullmq.NewQueue[PayoutJobData](PayoutQueueName, queueClient, &bullmq.QueueOptions{})
	if err != nil {
		return nil, fmt.Errorf("create payout queue: %w", err)
	}

	return &Service{
		repo:          repo,
		ledgerService: ledgerService,
		logger:        logger,
		queue:         queue,
	}, nil
}
func (s *Service) EnqueuePayout(ctx context.Context, payoutID uuid.UUID) error {
	jobData := PayoutJobData{PayoutID: payoutID.String()}

	_, err := s.queue.Add(
		ctx,
		"process-payout",
		jobData,
		bullmq.AddWithAttempts(3),
		bullmq.AddWithBackoff(bullmq.BackoffOptions{
			Type:  "exponential",
			Delay: 2000,
		}),
		bullmq.AddWithRemoveOnComplete(bullmq.KeepJobs{Count: 1000}),
		bullmq.AddWithRemoveOnFail(bullmq.KeepJobs{Count: 5000}),
	)
	if err != nil {
		return fmt.Errorf("enqueue payout job: %w", err)
	}

	s.logger.Info("payout job enqueued", "payout_id", payoutID)
	return nil
}
func (s *Service) StartWorker(ctx context.Context, workerClient redis.Cmdable) error {
	process := func(jobCtx context.Context, job *bullmq.Job[PayoutJobData]) (map[string]string, error) {
		data := job.Data()
		payoutID, err := uuid.Parse(data.PayoutID)
		if err != nil {
			return nil, fmt.Errorf("parse payout ID: %w", err)
		}

		s.logger.Info("worker picked up payout job", "payout_id", payoutID, "job_id", job.ID())
		if err := s.transitionToProcessing(ctx, payoutID); err != nil {
			return nil, fmt.Errorf("transition to processing: %w", err)
		}
		time.Sleep(3 * time.Second)
		if rand.Float64() < 0.8 {
			s.completeSuccess(ctx, payoutID)
		} else {
			s.completeFailed(ctx, payoutID, "Simulated bank rejection: account not found")
		}

		return map[string]string{"payout_id": data.PayoutID, "status": "resolved"}, nil
	}

	worker, err := bullmq.NewWorker[PayoutJobData, map[string]string](PayoutQueueName, workerClient, process, &bullmq.WorkerOptions{
		Concurrency:     5,
		StalledInterval: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("create payout worker: %w", err)
	}

	s.logger.Info("payout worker started", "queue", PayoutQueueName, "concurrency", 5)
	return worker.Run(ctx)
}
func (s *Service) transitionToProcessing(ctx context.Context, payoutID uuid.UUID) error {
	tx, err := s.repo.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := s.repo.UpdatePayoutStatus(ctx, tx, payoutID, models.PayoutProcessing, nil); err != nil {
		tx.Rollback()
		return fmt.Errorf("set processing: %w", err)
	}

	payout, err := s.repo.GetPayoutByID(ctx, payoutID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("get payout: %w", err)
	}

	if err := s.repo.UpdateTransactionStatus(ctx, tx, payout.TransactionID, models.StatusProcessing); err != nil {
		tx.Rollback()
		return fmt.Errorf("update tx status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	s.logger.Info("payout moved to processing", "payout_id", payoutID)
	return nil
}

func (s *Service) completeSuccess(ctx context.Context, payoutID uuid.UUID) {
	tx, err := s.repo.DB().BeginTxx(ctx, nil)
	if err != nil {
		s.logger.Error("payout success: begin tx", "error", err)
		return
	}

	if err := s.repo.UpdatePayoutStatus(ctx, tx, payoutID, models.PayoutSuccessful, nil); err != nil {
		tx.Rollback()
		s.logger.Error("payout success: update status", "error", err)
		return
	}

	payout, _ := s.repo.GetPayoutByID(ctx, payoutID)
	if payout != nil {
		s.repo.UpdateTransactionStatus(ctx, tx, payout.TransactionID, models.StatusSuccessful)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("payout success: commit", "error", err)
		return
	}

	s.logger.Info("payout completed successfully", "payout_id", payoutID)
}

func (s *Service) completeFailed(ctx context.Context, payoutID uuid.UUID, reason string) {
	payout, err := s.repo.GetPayoutByID(ctx, payoutID)
	if err != nil {
		s.logger.Error("payout failure: get payout", "error", err)
		return
	}

	tx, err := s.repo.DB().BeginTxx(ctx, nil)
	if err != nil {
		s.logger.Error("payout failure: begin tx", "error", err)
		return
	}

	if err := s.repo.UpdatePayoutStatus(ctx, tx, payoutID, models.PayoutFailed, &reason); err != nil {
		tx.Rollback()
		s.logger.Error("payout failure: update status", "error", err)
		return
	}

	if err := s.repo.UpdateTransactionStatus(ctx, tx, payout.TransactionID, models.StatusFailed); err != nil {
		tx.Rollback()
		s.logger.Error("payout failure: update tx status", "error", err)
		return
	}

	wallet, err := s.repo.GetWalletByUserIDTx(ctx, tx, payout.UserID)
	if err != nil {
		tx.Rollback()
		s.logger.Error("payout failure: get wallet", "error", err)
		return
	}

	reversalMeta := map[string]interface{}{
		"original_payout_id":      payout.ID,
		"original_transaction_id": payout.TransactionID,
		"reason":                  reason,
	}

	reversalTx, err := s.repo.CreateTransaction(ctx, tx, payout.UserID,
		models.TransactionPayout, models.StatusSuccessful, payout.SourceCurrency, payout.Amount, nil, reversalMeta)
	if err != nil {
		tx.Rollback()
		s.logger.Error("payout failure: create reversal tx", "error", err)
		return
	}

	if err := s.ledgerService.RecordPayoutReversal(ctx, tx, reversalTx.ID, wallet.ID, payout.SourceCurrency, payout.Amount); err != nil {
		tx.Rollback()
		s.logger.Error("payout failure: write reversal ledger", "error", err)
		return
	}

	if err := s.repo.SetPayoutReversalTx(ctx, tx, payoutID, reversalTx.ID); err != nil {
		tx.Rollback()
		s.logger.Error("payout failure: link reversal", "error", err)
		return
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("payout failure: commit reversal", "error", err)
		return
	}

	s.logger.Info("payout failed and reversed", "payout_id", payoutID, "reason", reason)
}
func (s *Service) ManualTransition(ctx context.Context, payoutID uuid.UUID, targetStatus models.PayoutStatus, reason *string) error {
	payout, err := s.repo.GetPayoutByID(ctx, payoutID)
	if err != nil {
		return fmt.Errorf("get payout: %w", err)
	}

	validTransitions := map[models.PayoutStatus][]models.PayoutStatus{
		models.PayoutPending:    {models.PayoutProcessing, models.PayoutFailed},
		models.PayoutProcessing: {models.PayoutSuccessful, models.PayoutFailed},
	}

	allowed, ok := validTransitions[payout.Status]
	if !ok {
		return fmt.Errorf("payout in terminal state: %s", payout.Status)
	}

	valid := false
	for _, s := range allowed {
		if s == targetStatus {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid transition: %s → %s", payout.Status, targetStatus)
	}

	if targetStatus == models.PayoutFailed {
		s.completeFailed(ctx, payoutID, *reason)
		return nil
	}

	if targetStatus == models.PayoutSuccessful {
		s.completeSuccess(ctx, payoutID)
		return nil
	}

	tx, err := s.repo.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePayoutStatus(ctx, tx, payoutID, targetStatus, reason); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}