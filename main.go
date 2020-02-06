package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/BurntSushi/toml"
	cron "github.com/robfig/cron/v3"
)

type job struct {
	Command  string        `toml:"command"`
	Crontime string        `toml:"crontime"`
	Timeout  time.Duration `toml:"timeout"`
}

type jobs struct {
	Jobs []job `toml:"job"`
}

func main() {
	l := cron.VerbosePrintfLogger(log.New(os.Stdout, "", log.LstdFlags))

	var jj jobs

	if _, err := toml.DecodeFile("./jobs.toml", &jj); err != nil {
		l.Error(err, "error reading file")
		os.Exit(1)
	}

	if len(jj.Jobs) == 0 {
		l.Info("no jobs. exiting...")
		os.Exit(0)
	}

	c := cron.New(cron.WithLogger(l))

	for _, j := range jj.Jobs {
		kv := []interface{}{
			"command",
			j.Command,
			"crontime",
			j.Crontime,
			"timeout",
			j.Timeout,
		}

		_, err := c.AddFunc(j.Crontime, func() {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, j.Timeout)

			if err := exec.CommandContext(ctx, j.Command).Run(); err != nil {
				l.Error(err, "error running command", kv...)
			}

			cancel()
		})

		if err != nil {
			l.Error(err, "function add error", kv...)
			continue
		}
	}

	c.Run()
}
