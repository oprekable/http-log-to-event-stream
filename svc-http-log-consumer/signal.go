package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

var (
	ErrTermSig = errors.New("termination signal caught")
)

type SignalTrap chan os.Signal

func TermSignalTrap() SignalTrap {
	trap := SignalTrap(make(chan os.Signal, 1))
	signal.Notify(trap, syscall.SIGINT, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSEGV)

	return trap
}

func (t SignalTrap) Wait(ctx context.Context) error {
	select {
	case <-t:
		return ErrTermSig
	case <-ctx.Done():
		return ctx.Err()
	}
}
