package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	app "github.com/parabrola/cook/internal"
)

func main() {
	opts := app.NewCliOptions()

	handleGlobalOptions(&opts, nil)

	cfg, err := app.ReadConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fs := app.LocalFileSystem{}
	proc := app.ShellProcess{}

	p := app.NewParser(cfg, &opts, &fs)
	if err := p.Bootstrap(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	handleGlobalOptions(&opts, &p)

	l := app.NewLockfile(p.GetFilePaths(), &opts, &fs)
	if err := l.Bootstrap(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	e := app.NewExecutor(&p, &l, &opts, &proc, &fs, &ctx)
	e.Start(opts.TaskName)
}
