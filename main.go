package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/po3rin/esazoth/es"
	"github.com/sourcegraph/conc/pool"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

var rootCmd = &cobra.Command{
	Use:   "esazoth",
	Short: "esazoth recives reindex task ID and wait completed and returns the recommended document sync batch period",
	Long:  "esazoth recives reindex task ID and wait completed and returns the recommended document sync batch period",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

		err := run()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	},
}

func run() error {
	es, err := es.NewClient()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	taskID := scanner.Text()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := pool.New().WithErrors().WithContext(ctx)
	pool.Go(func(ctx context.Context) error {
		defer cancel()
		for {
			select {
			case <-time.Tick(3 * time.Second):
				esResTask, err := es.Task(ctx, taskID)
				if err != nil {
					return err
				}

				if esResTask.Completed {
					if esResTask.Error == nil {
						if !esResTask.Response.TimeOut {
							d := time.Unix(int64(esResTask.Task.StartTimeInMillis/1000), 0)
							s := time.Since(d)
							result := int64(s.Hours() / 24)
							fmt.Println(result)
							return nil
						}
						return fmt.Errorf("task(%v) is timeout", esResTask.Task.ID)
					}
					return fmt.Errorf("task(%v) has error: type: %v, reason: %v", esResTask.Task.ID, esResTask.Error.Type, esResTask.Error.Reason)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		cancel()
	case <-ctx.Done():
	}

	if err := pool.Wait(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
