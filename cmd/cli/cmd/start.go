package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tuongaz/go-saas/api"
	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/pkg/log"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start service",
	Long:  `Start service`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustCreateConfig()
		api := api.New(cfg)

		log.Error("API stopped", api.Start(context.Background()))
	},
}

func mustCreateConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		log.Default().Error("failed to init config", log.ErrorAttr(err))
		panic(err)
	}

	return cfg
}
