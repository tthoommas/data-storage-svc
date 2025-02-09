package endpoints

import (
	"data-storage-svc/internal"
	"data-storage-svc/internal/api/common"
	"data-storage-svc/internal/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserEndpoint interface {
	common.EndpointGroup
	// Create (register) a new user
	Create(c *gin.Context)
	// Fetch a JWT token to authenticate user
	FetchToken(c *gin.Context)
	// Logout, i.e. delete the JWT token
	Logout(c *gin.Context)
}
type userEndpoint struct {
	common.EndpointGroup
	userService services.UserService
}

func NewUserEndpoint(
	// Common dependencies
	commonMiddlewares []gin.HandlerFunc,
	permissionsManager common.PermissionsManager,
	// Service dependencies
	userService services.UserService,
) UserEndpoint {
	userEndpoint := userEndpoint{userService: userService}

	endpoint := common.NewEndpoint(
		"Users",
		"/user",
		commonMiddlewares,
		map[common.MethodPath][]gin.HandlerFunc{
			{Method: "POST", Path: ""}:        {userEndpoint.Create},
			{Method: "POST", Path: "/jwt"}:    {userEndpoint.FetchToken},
			{Method: "POST", Path: "/logout"}: {userEndpoint.Logout},
		},
		permissionsManager,
	)

	userEndpoint.EndpointGroup = endpoint
	return &userEndpoint
}

type RegisterUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (e *userEndpoint) Create(c *gin.Context) {
	var registerUserBody RegisterUserBody

	if err := c.BindJSON(&registerUserBody); err != nil {
		return
	}

	_, svcErr := e.userService.Create(registerUserBody.Email, registerUserBody.Password)
	if svcErr != nil {
		svcErr.Apply(c)
		return
	}

	c.Status(http.StatusCreated)
}

type FetchJWTBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (e *userEndpoint) FetchToken(c *gin.Context) {
	var fetchJWTBody FetchJWTBody

	if err := c.BindJSON(&fetchJWTBody); err != nil {
		return
	}

	jwt, err := e.userService.GenerateToken(fetchJWTBody.Email, fetchJWTBody.Password)
	if err != nil {
		err.Apply(c)
		return
	}

	c.SetCookie("jwt", *jwt, 3600, "/", internal.API_DOMAIN, !internal.DEBUG, true)
	c.SetCookie("user", fetchJWTBody.Email, 3600, "/", internal.API_DOMAIN, !internal.DEBUG, false)
	c.JSON(http.StatusOK, gin.H{"jwt": jwt})
}

func (e *userEndpoint) Logout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", internal.API_DOMAIN, !internal.DEBUG, true)
	c.SetCookie("user", "", -1, "/", internal.API_DOMAIN, !internal.DEBUG, false)
	c.Status(http.StatusOK)
}
