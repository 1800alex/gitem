package gitm

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func GetOSContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGINT and SIGTERM.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		cancel()
	}()

	return ctx, cancel
}
