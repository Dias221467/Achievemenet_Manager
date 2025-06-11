package middleware

import (
	"net/http"

	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateLastActiveMiddleware(userService *services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			if claims != nil {
				userID, err := primitive.ObjectIDFromHex(claims.UserID)
				if err == nil {
					_ = userService.UpdateLastActive(r.Context(), userID)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
