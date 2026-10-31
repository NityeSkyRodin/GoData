package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error with 'git %v': %s (%v)", args, stderr.String(), err)
	}
	return nil
}

func CommitRaceResult(trackName string, sessionType string) {

    if err := runGitCommand("add", "."); err != nil {
        panic(err)
    }

    commitMsg := fmt.Sprintf("Added race result for %s - %s", trackName, sessionType)
    if err := runGitCommand("commit", "-m", commitMsg); err != nil {
        panic(err)
    }

    if err := runGitCommand("push"); err != nil {
        panic(err)
    }

    fmt.Println("Succesfully pushed")
}