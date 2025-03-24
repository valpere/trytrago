package docs

import (
	"github.com/gin-gonic/gin"
)

// RegisterSwaggerEndpoints registers the Swagger UI endpoints
func RegisterSwaggerEndpoints(router *gin.Engine) {
	// Serve the OpenAPI specification file
	router.GET("/v3/api-docs", func(c *gin.Context) {
		c.File("./interface/api/rest/docs/openapi.yaml")
	})

	// Serve the Swagger UI HTML page directly
	router.GET("/swagger-ui.html", func(c *gin.Context) {
		c.File("./interface/api/rest/docs/swagger-ui/index.html")
	})

	// Serve Swagger UI static files
	router.Static("/swagger-ui", "./interface/api/rest/docs/swagger-ui")
}
