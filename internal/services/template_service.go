package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemplateService struct {
	repo     *repository.TemplateRepository
	goalRepo *repository.GoalRepository
}

func NewTemplateService(repo *repository.TemplateRepository, goalRepo *repository.GoalRepository) *TemplateService {
	return &TemplateService{
		repo:     repo,
		goalRepo: goalRepo,
	}
}

// CreateTemplate creates a new goal template
func (s *TemplateService) CreateTemplate(ctx context.Context, template *models.GoalTemplate) (*models.GoalTemplate, error) {
	if template.Title == "" || len(template.Steps) == 0 {
		return nil, fmt.Errorf("template must have a title and at least one step")
	}
	return s.repo.CreateTemplate(ctx, template)
}

// GetAllTemplates returns all available templates
func (s *TemplateService) GetAllTemplates(ctx context.Context) ([]models.GoalTemplate, error) {
	return s.repo.GetAllTemplates(ctx)
}

// GetTemplateByID retrieves a single template by ID
func (s *TemplateService) GetTemplateByID(ctx context.Context, id string) (*models.GoalTemplate, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID")
	}
	return s.repo.GetTemplateByID(ctx, objID)
}

func (s *TemplateService) CopyTemplateToGoal(ctx context.Context, templateID string, userID primitive.ObjectID) (*models.Goal, error) {
	objID, err := primitive.ObjectIDFromHex(templateID)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID")
	}

	template, err := s.repo.GetTemplateByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %v", err)
	}

	var steps []models.Step
	for _, tmplStep := range template.Steps {
		var substeps []models.Substep
		for _, tmplSub := range tmplStep.Substeps {
			substeps = append(substeps, models.Substep{
				Title: tmplSub.Title,
				Done:  false,
			})
		}
		steps = append(steps, models.Step{
			Name:      tmplStep.Name,
			Substeps:  substeps,
			Completed: false,
		})
	}

	goal := &models.Goal{
		Name:        template.Title,
		Description: template.Description,
		Steps:       steps, // Previously converted from []string
		Category:    template.Category,
		UserID:      userID,
		Status:      "in_progress",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.goalRepo.CreateGoal(ctx, goal)
}

func (s *TemplateService) GetTemplatesByUser(ctx context.Context, userID primitive.ObjectID) ([]models.GoalTemplate, error) {
	return s.repo.GetTemplatesByUser(ctx, userID)
}

func (s *TemplateService) GetPublicTemplates(ctx context.Context) ([]models.GoalTemplate, error) {
	return s.repo.GetPublicTemplates(ctx)
}

func (s *TemplateService) GetPublicTemplatesByUser(ctx context.Context, userID primitive.ObjectID) ([]models.GoalTemplate, error) {
	return s.repo.GetPublicTemplatesByUser(ctx, userID)
}
