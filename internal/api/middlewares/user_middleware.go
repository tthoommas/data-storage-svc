package middlewares

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/security"
	"data-storage-svc/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// Decode a fetch user in DB if any, do NOT reject the request if no user is found
func UserMiddleware(userRepository repository.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Check if a jwt is embedded in the request
		authCookie, err := ctx.Cookie("jwt")
		if err != nil || authCookie == "" {
			ctx.Next()
			return
		}

		// Parse the jwt
		token, err := jwt.Parse(authCookie, func(token *jwt.Token) (interface{}, error) {
			return security.GetSecretKey(), nil
		})

		if err != nil || !token.Valid {
			// The token is not valid, no user to get
			ctx.Next()
			return
		}

		// JWT is valid, extract email claim from the token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			email := claims["email"].(string)
			user, err := userRepository.GetByEmail(email)
			if err != nil {
				// Couldn't get the user from the JWT, maybe it was delete recently, after the token was generated
				ctx.Next()
				return
			}
			// Store the user in context
			ctx.Set(common.USER, user)
		}
		ctx.Next()
	}
}
