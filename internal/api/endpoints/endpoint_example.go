package endpoints

import (
	"data-storage-svc/internal/api/common"

	"github.com/gin-gonic/gin"
)

// This is an example endpoint template
type MyEndpoint interface {
	common.EndpointGroup
	MyPOST(c *gin.Context)
	MyGET(c *gin.Context)
}

type myEndpoint struct {
	common.EndpointGroup
	// Service dependencies
}

func NewMyEndpoint(
	// Common dependencies
	commonMiddlewares []gin.HandlerFunc,
	permissionsManager common.PermissionsManager,
	//Services dependencies
) MyEndpoint {

	myEndpoint := myEndpoint{}

	endpoint := common.NewEndpoint(
		"Example",
		"/example",
		commonMiddlewares,
		map[common.MethodPath][]gin.HandlerFunc{
			// Common album edition actions
			{Method: "POST", Path: "/"}: {myEndpoint.MyPOST},
			{Method: "GET", Path: "/"}:  {myEndpoint.MyGET},
		},
		permissionsManager,
	)

	myEndpoint.EndpointGroup = endpoint
	return myEndpoint
}

func (e myEndpoint) MyPOST(c *gin.Context) {

}

func (e myEndpoint) MyGET(c *gin.Context) {

}
