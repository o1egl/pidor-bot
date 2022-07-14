package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/o1egl/pidor-bot/cmd/run"
	"github.com/o1egl/pidor-bot/config"
	"github.com/o1egl/pidor-bot/repo"
	"github.com/o1egl/pidor-bot/service/pidor"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "start bot",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Failed to load config", err)
			os.Exit(1)
		}

		logger, err := newLogger()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		db, err := bbolt.Open(cfg.DBPath, 0600, nil)
		if err != nil {
			logger.Fatal("Failed to open database", zap.Error(err))
		}

		repoClient := repo.NewBoltRepo(db)

		pidorService, err := pidor.New(cfg, logger, repoClient)
		if err != nil {
			logger.Fatal("Failed to initialize pidor service", zap.Error(err))
		}

		app := run.NewApp(logger, []run.Service{pidorService})
		if err := app.Start(); err != nil {
			logger.Error(err.Error())
		}
	},
}
