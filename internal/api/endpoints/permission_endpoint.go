package endpoints

import (
	"data-storage-svc/internal/database"
	"data-storage-svc/internal/model"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GrantGlobalPermissionBody struct {
	Email      *string `json:"targetUser"`
	Permission *int    `json:"grantedPermission"`
}

func GrantGlobalPermission(c *gin.Context) {
	var grantGlobalPermissionBody GrantGlobalPermissionBody

	if err := c.BindJSON(&grantGlobalPermissionBody); err != nil {
		return
	}

	initiator, _ := c.Get("user")
	initiatorUser := initiator.(*model.User)
	targetUser, err := database.FindUserByEmail(grantGlobalPermissionBody.Email)
	if err != nil {
		slog.Debug("couldn't grant permission to user", "error", err, "targetUserEmail", *grantGlobalPermissionBody.Email, "initatorEmail", *initiatorUser.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "couldn't grant permission"})
		return
	}

	if database.GrantGlobalPermission(targetUser.Email, grantGlobalPermissionBody.Permission) {
		slog.Debug("Successfully granted permission", "targetUserEmail", *grantGlobalPermissionBody.Email, "permission", *grantGlobalPermissionBody.Permission, "initatorEmail", *initiatorUser.Email)
		c.Status(http.StatusOK)
		return
	} else {
		slog.Debug("Couldn't grant permission", "targetUserEmail", *grantGlobalPermissionBody.Email, "permission", *grantGlobalPermissionBody.Permission, "initatorEmail", *initiatorUser.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "couldn't grant permission"})
		return
	}
}

type RevokeGlobalPermissionBody struct {
	Email      *string `json:"targetUser"`
	Permission *int    `json:"revokedPermission"`
}

func RevokeGlobalPermission(c *gin.Context) {
	var revokeGlobalPermissionBody RevokeGlobalPermissionBody

	if err := c.BindJSON(&revokeGlobalPermissionBody); err != nil {
		return
	}

	initiator, _ := c.Get("user")
	initiatorUser := initiator.(*model.User)
	targetUser, err := database.FindUserByEmail(revokeGlobalPermissionBody.Email)
	if err != nil {
		slog.Debug("couldn't revoke permission to user", "error", err, "targetUserEmail", *revokeGlobalPermissionBody.Email, "initatorEmail", *initiatorUser.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "couldn't grant permission"})
		return
	}

	if database.RevokeGlobalPermission(targetUser.Email, revokeGlobalPermissionBody.Permission) {
		slog.Debug("Successfully revoked permission", "targetUserEmail", *revokeGlobalPermissionBody.Email, "permission", *revokeGlobalPermissionBody.Permission, "initatorEmail", *initiatorUser.Email)
		c.Status(http.StatusOK)
		return
	} else {
		slog.Debug("Couldn't revoke permission", "targetUserEmail", *revokeGlobalPermissionBody.Email, "permission", *revokeGlobalPermissionBody.Permission, "initatorEmail", *initiatorUser.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "couldn't revoke permission"})
		return
	}
}
