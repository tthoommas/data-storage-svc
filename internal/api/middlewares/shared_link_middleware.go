package middlewares

import (
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/repository"

	"github.com/gin-gonic/gin"
)

func SharedLinkMiddleware(sharedLinkRepository repository.SharedLinkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if we have a token in the request
		sharedLinkToken := c.Query("token")
		if sharedLinkToken != "" {
			link, err := sharedLinkRepository.GetByToken(sharedLinkToken)
			// Add it to the context if any
			if err == nil && link != nil {
				c.Set(common.SHARED_LINK, link)
			}
		}
	}
}
