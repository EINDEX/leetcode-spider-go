package utils

import (
	"fmt"
	"leetcode-spider-go/settings"
	"log"
	"os"
	"os/exec"
	"time"
)

var Git git

type git struct {
}

func (git *git) execCommand(args ...string) (err error) {
	if !settings.Setting.EnableGit {
		return
	}
	log.Println("run command:", "git", args)
	cmd := exec.Command("git", args...)
	cmd.Dir = settings.Setting.Out
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	return
}

func GitAddAndCommand(path string, commitMessage string, timestamp int64) {
	if err := Git.execCommand("add", path); err != nil {
		log.Fatalln(err)
	}
	if err := Git.execCommand("commit", "--date", time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"), "-m", commitMessage); err != nil {
		log.Fatalln(err)
	}
}

func GitPush() {
	if !settings.Setting.EnablePush {
		return
	}
	if err := Git.execCommand("push"); err != nil {
		log.Fatalln(err)
	}
}
