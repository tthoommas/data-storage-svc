package utils

import "github.com/gin-gonic/gin"

type ServiceError interface {
	Apply(c *gin.Context)
	GetCode() int
}

type serviceError struct {
	HttpReturnCode    int
	HttpReturnMessage string
}

func NewServiceError(code int, message string) serviceError {
	return serviceError{code, message}
}

func (s serviceError) Apply(c *gin.Context) {
	c.IndentedJSON(s.HttpReturnCode, gin.H{"error": s.HttpReturnMessage})
}

func (s serviceError) GetCode() int {
	return s.HttpReturnCode
}
