package main

import (
	"context"

	_ "github.com/th1enq/es-demo/docs" // Import swagger docs
	"github.com/th1enq/es-demo/internal/app"
)

// @title           Event Sourcing Demo API
// @version         1.0
// @description     Event Sourcing Demo API with Bank Account CQRS Pattern
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @schemes http https
func main() {
	app, err := app.Initialize(context.Background())
	if err != nil {
		panic("Failed to initialize application: " + err.Error())
	}
	app.Start(context.Background())
}
