package lambda

import (
	"bufio"
	"context"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/robfig/cron"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

var targetsPattern = regexp.MustCompile(`^([\d\w-/]+)\s*:\s*[\d\w-/\s]*$`)

// List Make actions (if Makefile defined)
func (local *localLambda) Actions() ([]string, error) {
	makefile := filepath.Join(local.rootDir, "Makefile")
	f, err := os.Open(makefile)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	var ans = make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		matches := targetsPattern.FindAllStringSubmatch(scanner.Text(), -1)
		if len(matches) < 1 || len(matches[0]) != 2 {
			continue
		}
		ans = append(ans, matches[0][1])
	}
	return ans, nil
}

// Invoke action by name (make target)
func (local *localLambda) Do(ctx context.Context, name string, timeLimit time.Duration, globalEnv map[string]string, out io.Writer) error {
	if out == nil {
		out = os.Stderr
	}
	if timeLimit > 0 {
		cctx, cancel := context.WithTimeout(ctx, timeLimit)
		defer cancel()
		ctx = cctx
	}
	environments := os.Environ()
	for k, v := range globalEnv {
		environments = append(environments, k+"="+v)
	}
	for k, v := range local.manifest.Environment {
		environments = append(environments, k+"="+v)
	}

	cmd := exec.CommandContext(ctx, "make", name)
	cmd.Dir = local.rootDir
	cmd.Stdout = out
	cmd.Stderr = out
	internal.SetCreds(cmd, local.creds)
	internal.SetFlags(cmd)
	cmd.Env = environments

	return cmd.Run()
}

func (local *localLambda) DoScheduled(ctx context.Context, lastRun time.Time, globalEnv map[string]string) {
	now := time.Now()
	for _, plan := range local.manifest.Cron {
		sched, err := cron.Parse(plan.Cron)
		if err != nil {
			log.Println(plan.Cron, "-", err)
			continue
		}
		if !sched.Next(lastRun).After(now) {
			err = local.Do(ctx, plan.Action, time.Duration(plan.TimeLimit), globalEnv, nil)
			if err != nil {
				log.Println(plan.Cron, plan.Action, err)
			}
		}
	}
}
