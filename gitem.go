package main

import (
	"context"
	"sync"

	"github.com/fatih/color"
)

type Gitem struct {
	ctx    context.Context
	cancel context.CancelFunc

	config Config

	printMu sync.Mutex

	wg sync.WaitGroup

	Info    *color.Color
	Warning *color.Color
	Error   *color.Color
}

func New() *Gitem {
	var g Gitem

	g.Error = color.New(color.FgRed, color.Bold)
	g.Info = color.New(color.FgBlue, color.Bold)
	g.Warning = color.New(color.FgYellow, color.Bold)

	g.ctx, g.cancel = context.WithCancel(context.Background())

	return &g
}
