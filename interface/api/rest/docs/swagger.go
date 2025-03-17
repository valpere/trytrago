// interface/api/rest/docs/swagger.go

package docs

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed swagger-ui
var swaggerUIFiles embed.FS

//go:embed openapi.yaml
var openAPISpec []byte

// RegisterSwaggerEndpoints registers the Swagger UI and OpenAPI specification endpoints
func RegisterSwaggerEndpoints(router *gin.Engine) {
	// Serve OpenAPI specification
	router.GET("/v3/api-docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", openAPISpec)
	})

	// Redirect /swagger-ui.html to /swagger-ui/
	router.GET("/swagger-ui.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger-ui/")
	})

	// Serve Swagger UI files
	subFS, err := fs.Sub(swaggerUIFiles, "swagger-ui")
	if err != nil {
		panic(err)
	}

	router.GET("/swagger-ui/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		// Default to index.html for the root path
		if path == "/" || path == "" {
			path = "/index.html"
		}

		// Remove leading slash to match filesystem paths
		path = strings.TrimPrefix(path, "/")

		// Special handling for index.html to configure OpenAPI URL
		if path == "index.html" {
			content, err := fs.ReadFile(subFS, path)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}

			// Replace the default URL with our API docs URL
			modifiedContent := strings.Replace(
				string(content),
				"https://petstore.swagger.io/v2/swagger.json",
				"/v3/api-docs",
				1,
			)

			c.Data(http.StatusOK, "text/html", []byte(modifiedContent))
			return
		}

		// Serve other static files directly
		c.FileFromFS(path, http.FS(subFS))
	})
}
