package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	jwtutil "github.com/Dias221467/Achievemenet_Manager/pkg/jwt"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type WSMessage struct {
	Type       string `json:"type"`        // "text", "status", "typing", "file", "image", "audio"
	ReceiverID string `json:"receiver_id"`
	SenderID   string `json:"sender_id,omitempty"`
	Text       string `json:"text,omitempty"`
	FileURL    string `json:"fileUrl,omitempty"`
	FileName   string `json:"fileName,omitempty"`
	Typing     bool   `json:"typing,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

type ChatHandler struct {
	Service   *services.ChatService
	JWTSecret string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}


var (
	clients   = make(map[string]*websocket.Conn)
	clientsMu sync.Mutex
	online    = make(map[string]bool)
)

func broadcastStatus(userID, status string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, conn := range clients {
		conn.WriteJSON(map[string]interface{}{
			"type":   "status",
			"userId": userID,
			"status": status, // "online" или "offline"
		})
	}
}

func NewChatHandler(service *services.ChatService, jwtSecret string) *ChatHandler {
	return &ChatHandler{Service: service, JWTSecret: jwtSecret}
}

// ======== WebSocket Chat ========
func (h *ChatHandler) ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}
	claims, err := jwtutil.ValidateToken(token, h.JWTSecret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		fmt.Println("WebSocket auth failed:", err)
		return
	}
	userID := claims.UserID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	fmt.Println("WebSocket connected:", userID)

	clientsMu.Lock()
	clients[userID] = conn
	online[userID] = true
	clientsMu.Unlock()
	broadcastStatus(userID, "online")

	defer func() {
		clientsMu.Lock()
		delete(clients, userID)
		delete(online, userID)
		clientsMu.Unlock()
		broadcastStatus(userID, "offline")
		conn.Close()
		fmt.Println("WebSocket disconnected:", userID)
	}()

	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			break // клиент отключился
		}

	
		if msg.Type == "typing" {
			clientsMu.Lock()
			if receiverConn, ok := clients[msg.ReceiverID]; ok {
				receiverConn.WriteJSON(map[string]interface{}{
					"type":      "typing",
					"sender_id": userID,
					"typing":    msg.Typing,
				})
			}
			clientsMu.Unlock()
			continue
		}

	
		if (msg.Type == "file" || msg.Type == "image" || msg.Type == "audio") && msg.FileURL != "" {
			senderObjID, _ := primitive.ObjectIDFromHex(userID)
			receiverObjID, _ := primitive.ObjectIDFromHex(msg.ReceiverID)
			newMsg := &models.Message{
				SenderID:   senderObjID,
				ReceiverID: receiverObjID,
				Type:       msg.Type,
				FileURL:    msg.FileURL,
				FileName:   msg.FileName,
				CreatedAt:  time.Now(),
			}
			h.Service.SendMessage(r.Context(), newMsg)

			clientsMu.Lock()
			response := map[string]interface{}{
				"type":        msg.Type,
				"sender_id":   userID,
				"receiver_id": msg.ReceiverID,
				"file_url":    msg.FileURL,
				"file_name":   msg.FileName,
				"created_at":  newMsg.CreatedAt,
				"id":          newMsg.ID.Hex(),
			}
			if receiverConn, ok := clients[msg.ReceiverID]; ok {
				receiverConn.WriteJSON(response)
			}
			_ = conn.WriteJSON(response)
			clientsMu.Unlock()
			continue
		}

		
		if msg.Type == "" || msg.Type == "text" {
			senderObjID, _ := primitive.ObjectIDFromHex(userID)
			receiverObjID, _ := primitive.ObjectIDFromHex(msg.ReceiverID)
			newMsg := &models.Message{
				SenderID:   senderObjID,
				ReceiverID: receiverObjID,
				Type:       "text",
				Text:       msg.Text,
				CreatedAt:  time.Now(),
			}
			h.Service.SendMessage(r.Context(), newMsg)
			clientsMu.Lock()
			if receiverConn, ok := clients[msg.ReceiverID]; ok {
				_ = receiverConn.WriteJSON(map[string]interface{}{
					"type":        "text",
					"sender_id":   userID,
					"receiver_id": msg.ReceiverID,
					"text":        msg.Text,
					"created_at":  newMsg.CreatedAt,
				})
			}
			_ = conn.WriteJSON(map[string]interface{}{
				"type":        "text",
				"sender_id":   userID,
				"receiver_id": msg.ReceiverID,
				"text":        msg.Text,
				"created_at":  newMsg.CreatedAt,
			})
			clientsMu.Unlock()
		}
	}
}

func (h *ChatHandler) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*jwtutil.Claims)
	if !ok || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	friendId := vars["friendId"]

	messages, err := h.Service.GetChat(r.Context(), claims.UserID, friendId)
	if err != nil {
		http.Error(w, "Failed to get chat history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}


func (h *ChatHandler) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // max ~10MB
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)
	filePath := "./uploads/" + fileName

	out, err := createFile(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	fileUrl := "/uploads/" + fileName

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url":  fileUrl,
		"name": handler.Filename,
	})
}

func createFile(path string) (*os.File, error) {
	dir := "./uploads"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}
	return os.Create(path)
}
