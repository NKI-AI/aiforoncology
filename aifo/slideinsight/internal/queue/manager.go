// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package queue

import (
	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// QueueManager provides a high-level interface for managing queues
type QueueManager struct {
	queue        *Queue
	emailService ports.EmailService
}

// NewQueueManager creates a new queue manager
func NewQueueManager(config QueueConfig, emailService ports.EmailService) *QueueManager {
	return &QueueManager{
		queue:        NewQueue(config),
		emailService: emailService,
	}
}

// Start starts the queue manager
func (qm *QueueManager) Start() error {
	return qm.queue.Start()
}

// Stop stops the queue manager
func (qm *QueueManager) Stop() error {
	return qm.queue.Stop()
}

// SubmitEmailTask submits an email task to the queue
func (qm *QueueManager) SubmitEmailTask(request ports.EmailRequest) error {
	task := NewEmailTask(qm.emailService, request)
	return qm.queue.Submit(task)
}

// SubmitPasswordResetEmail submits a password reset email task
func (qm *QueueManager) SubmitPasswordResetEmail(email, token string) error {
	builder := NewEmailTaskBuilder(qm.emailService)
	task := builder.BuildPasswordResetTask(email, token)
	return qm.queue.Submit(task)
}

// SubmitEmailVerification submits an email verification task
func (qm *QueueManager) SubmitEmailVerification(email, token string) error {
	builder := NewEmailTaskBuilder(qm.emailService)
	task := builder.BuildEmailVerificationTask(email, token)
	return qm.queue.Submit(task)
}

// SubmitWelcomeEmail submits a welcome email task
func (qm *QueueManager) SubmitWelcomeEmail(email string) error {
	builder := NewEmailTaskBuilder(qm.emailService)
	task := builder.BuildWelcomeTask(email)
	return qm.queue.Submit(task)
}

// SubmitAlgorithmTask submits an algorithm task to the queue
func (qm *QueueManager) SubmitAlgorithmTask(name string, function AlgorithmFunction, parameters map[string]interface{}) error {
	task := NewAlgorithmTask(name, function, parameters)
	return qm.queue.Submit(task)
}

// Stats returns queue statistics
func (qm *QueueManager) Stats() map[string]interface{} {
	return qm.queue.Stats()
}

// IsRunning returns whether the queue is running
func (qm *QueueManager) IsRunning() bool {
	return qm.queue.IsRunning()
}
