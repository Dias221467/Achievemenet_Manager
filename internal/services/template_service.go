package services

import (
	"context"
	"fmt"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemplateService struct {
	repo *repository.TemplateRepository
}

func NewTemplateService(repo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
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

func (s *TemplateService) GetTemplatesByUser(ctx context.Context, userID primitive.ObjectID) ([]models.GoalTemplate, error) {
	return s.repo.GetTemplatesByUser(ctx, userID)
}
