package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// A simple middleware to decode the given path param ID from the request and put it in context.
// Fails with 404 if not found
func PathParamIdMiddleware(pathParamNames ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, pathParamName := range pathParamNames {
			pathParamValue := ctx.Param(pathParamName)
			if pathParamValue == "" {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
			id, err := primitive.ObjectIDFromHex(pathParamValue)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			ctx.Set(pathParamName, id)
		}
		ctx.Next()
	}
}
