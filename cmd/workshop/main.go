package main

import (
	"context"
	_ "workshop/docs"
	"workshop/internal/app"
	_ "workshop/internal/controller"
	_ "workshop/internal/controller/echo"

	"github.com/rs/zerolog/log"
)

// @title						Workshop Service
// @version					1.0
//
// @host						localhost:8080
// @BasePath					/api
//
// @securityDefinitions.apikey	DebugUserRoles
// @in							header
// @name						X-Debug-User-Roles
// @description				Force user roles if debug mode is enabled. Split by comma ',' for multiple roles.
//
// @securityDefinitions.apikey	DebugUserID
// @in							header
// @name						X-Debug-User-ID
// @description				Force user ID if debug mode is enabled. Required if Debug User Roles is provided.
func main() {
	ctx := context.TODO()

	a, err := app.New(ctx)
	if err != nil {
		log.Fatal().Msgf("Failed to create app: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatal().Msgf("Failed to run app: %v", err)
	}

	if err := a.Stop(ctx); err != nil {
		log.Fatal().Msgf("Failed to stop app: %v", err)
	}
}
