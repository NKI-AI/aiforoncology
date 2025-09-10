// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/gofiber/fiber/v2/log"
)

// emailService implements the EmailService interface
type emailService struct {
	config             config.EmailConfig
	sesClient          *ses.Client
	templateRepository ports.EmailTemplateRepository
}

// NewEmailService creates a new email service based on the configuration
func NewEmailService(ctx context.Context, emailConfig config.EmailConfig, templateRepository ports.EmailTemplateRepository) (ports.EmailService, error) {
	service := &emailService{
		config:             emailConfig,
		templateRepository: templateRepository,
	}

	// Initialize SES client if using SES provider
	if emailConfig.Provider == "ses" {
		cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(emailConfig.SES.Region))
		if err != nil {
			return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
		}
		service.sesClient = ses.NewFromConfig(cfg)
		log.Info("SES email service initialized", "region", emailConfig.SES.Region)
	} else {
		log.Info("Console email service initialized")
	}

	return service, nil
}

// SendEmail sends an email using the specified template and data
func (s *emailService) SendEmail(ctx context.Context, request ports.EmailRequest) error {
	// Extract tenant ID - first check EmailRequest.TenantID, then fallback to data or context
	var tenantID int
	if request.TenantID > 0 {
		tenantID = request.TenantID
	} else if tid, ok := request.Data["tenantId"].(int); ok {
		tenantID = tid
	} else if ctxTenantID := ctx.Value("tenantId"); ctxTenantID != nil {
		if tid, ok := ctxTenantID.(int); ok {
			tenantID = tid
		}
	}

	if tenantID == 0 {
		return fmt.Errorf("tenant ID is required for email template lookup")
	}

	// Get the email template from the database
	emailTemplate, err := s.templateRepository.GetEmailTemplateByType(ctx, tenantID, request.Template)
	if err != nil {
		return fmt.Errorf("failed to get email template: %w", err)
	}

	// Render the template with the provided data
	subject, err := s.executeTemplate(emailTemplate.Subject, request.Data)
	if err != nil {
		return fmt.Errorf("failed to render email subject: %w", err)
	}

	bodyText, err := s.executeTemplate(emailTemplate.BodyText, request.Data)
	if err != nil {
		return fmt.Errorf("failed to render email text body: %w", err)
	}

	bodyHTML, err := s.executeTemplate(emailTemplate.BodyHTML, request.Data)
	if err != nil {
		return fmt.Errorf("failed to render email HTML body: %w", err)
	}

	return s.sendEmail(ctx, request.To, subject, bodyText, bodyHTML)
}

// executeTemplate executes a template with the given data
func (s *emailService) executeTemplate(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// SendPasswordResetEmail sends a password reset email
func (s *emailService) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenantId").(int)
	if !ok {
		return fmt.Errorf("tenant ID is required for email template lookup")
	}

	data := map[string]interface{}{
		"token":    token,
		"email":    email,
		"tenantId": tenantID,
	}

	request := ports.EmailRequest{
		To:       email,
		Template: ports.EmailTemplateTypePasswordReset,
		Data:     data,
	}

	return s.SendEmail(ctx, request)
}

// SendEmailVerificationEmail sends an email verification email
func (s *emailService) SendEmailVerificationEmail(ctx context.Context, email, token string) error {
	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenantId").(int)
	if !ok {
		return fmt.Errorf("tenant ID is required for email template lookup")
	}

	data := map[string]interface{}{
		"token":    token,
		"email":    email,
		"tenantId": tenantID,
	}

	request := ports.EmailRequest{
		To:       email,
		Template: ports.EmailTemplateTypeEmailVerification,
		Data:     data,
	}

	return s.SendEmail(ctx, request)
}

// SendWelcomeEmail sends a welcome email to new users
func (s *emailService) SendWelcomeEmail(ctx context.Context, email string) error {
	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenantId").(int)
	if !ok {
		return fmt.Errorf("tenant ID is required for email template lookup")
	}

	data := map[string]interface{}{
		"email":    email,
		"tenantId": tenantID,
	}

	request := ports.EmailRequest{
		To:       email,
		Template: ports.EmailTemplateTypeWelcome,
		Data:     data,
	}

	return s.SendEmail(ctx, request)
}

// sendEmail sends an email using the configured provider
func (s *emailService) sendEmail(ctx context.Context, to, subject, bodyText, bodyHTML string) error {
	switch s.config.Provider {
	case "ses":
		return s.sendSESEmail(ctx, to, subject, bodyText, bodyHTML)
	case "console":
		return s.sendConsoleEmail(to, subject, bodyText, bodyHTML)
	default:
		return fmt.Errorf("unsupported email provider: %s", s.config.Provider)
	}
}

// sendSESEmail sends an email using AWS SES
func (s *emailService) sendSESEmail(ctx context.Context, to, subject, bodyText, bodyHTML string) error {
	if s.sesClient == nil {
		return fmt.Errorf("SES client not initialized")
	}

	fromAddress := s.config.SES.FromAddress
	if s.config.SES.FromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", s.config.SES.FromName, s.config.SES.FromAddress)
	}

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(subject),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(bodyText),
				},
				Html: &types.Content{
					Data: aws.String(bodyHTML),
				},
			},
		},
		Source: aws.String(fromAddress),
	}

	result, err := s.sesClient.SendEmail(ctx, input)
	if err != nil {
		log.Error("Failed to send email via SES", "error", err, "to", to, "subject", subject)
		return fmt.Errorf("failed to send email via SES: %w", err)
	}

	log.Info("Email sent successfully via SES",
		"to", to,
		"subject", subject,
		"message_id", aws.ToString(result.MessageId))

	return nil
}

// sendConsoleEmail prints email to console (for development/testing)
func (s *emailService) sendConsoleEmail(to, subject, bodyText, bodyHTML string) error {
	fmt.Printf("\n=== EMAIL (Console) ===\n")
	fmt.Printf("To: %s\n", to)
	fmt.Printf("From: %s <%s>\n", s.config.SES.FromName, s.config.SES.FromAddress)
	fmt.Printf("Subject: %s\n\n", subject)
	fmt.Printf("--- Text Body ---\n%s\n\n", bodyText)
	fmt.Printf("--- HTML Body ---\n%s\n", bodyHTML)
	fmt.Printf("======================\n\n")

	log.Info("Email sent to console", "to", to, "subject", subject)
	return nil
}
