package main

import (
	"context"
	"github.com/rs/zerolog/log"
	_ "workshop/docs"
	"workshop/internal/app"
	_ "workshop/internal/controller"
	_ "workshop/internal/controller/echo"
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
	a, err := app.New()
	if err != nil {
		log.Fatal().Msgf("Failed to create app: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatal().Msgf("Failed to run app: %v", err)
	}

	if err := a.Stop(context.Background()); err != nil {
		log.Fatal().Msgf("Failed to stop app: %v", err)
	}
}
