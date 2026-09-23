package worker

import (
	"fmt"
	"log"

	"github.com/ramesh3214/taskflow/model"
	"github.com/ramesh3214/taskflow/service"
)

type EmailWorker struct {
	emailService *service.EmailService
}

func NewEmailWorker(
	emailService *service.EmailService,
) *EmailWorker {
	return &EmailWorker{
		emailService: emailService,
	}
}

func (w *EmailWorker) SendWelcomeEmail(task *model.Task) error {

	subject := "Welcome to TaskFlow"

	body := fmt.Sprintf(
		"Hello,\n\nWelcome to TaskFlow!\n\nYour account has been created successfully.\n\nRegards,\nTaskFlow Team",
	)

	err := w.emailService.SendEmail(
		task.Email,
		subject,
		body,
	)

	if err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	log.Printf("Welcome email sent successfully to %s", task.Email)

	return nil
}