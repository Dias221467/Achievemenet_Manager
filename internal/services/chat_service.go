package services

import (
	"context"
	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatService struct {
	Repo *repository.ChatRepository
}

func NewChatService(repo *repository.ChatRepository) *ChatService {
	return &ChatService{Repo: repo}
}

func (s *ChatService) SendMessage(ctx context.Context, msg *models.Message) (*models.Message, error) {
	return s.Repo.SendMessage(ctx, msg)
}

func (s *ChatService) GetChat(ctx context.Context, userID, friendID string) ([]models.Message, error) {
	uid, _ := primitive.ObjectIDFromHex(userID)
	fid, _ := primitive.ObjectIDFromHex(friendID)
	return s.Repo.GetChat(ctx, uid, fid)
}
