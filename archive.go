package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	_ "golang.org/x/tools/txtar"
)

func main() {
	goMod, err := ioutil.ReadFile("go.mod")
	if err != nil {
		log.Fatal(err)
	}
	modSum := fmt.Sprintf("%x", sha256.Sum256(goMod))
	ref := fmt.Sprintf("archive/gomod/%v", modSum)

	lsRemote, err := exec.Command("git", "ls-remote").CombinedOutput()
	if err != nil {
		log.Fatalf("git ls-remote: %v, %s", err, lsRemote)
	}
	if bytes.Contains(lsRemote, []byte(ref)) {
		// A ref already exists, so go.mod hasn't changed and
		// it's already archived. This is the common case.
		log.Printf("go.mod %v already archived as %v", modSum, ref)
		return
	}
	run(exec.Command("go", "mod", "vendor"))
	if _, err := os.Stat("vendor"); os.IsNotExist(err) {
		out, err := exec.Command("go", "mod", "graph").Output()
		if err != nil {
			log.Fatalf("go mod graph: %v", err)
		}
		if len(out) == 0 {
			log.Printf("no deps; nothing to do")
			return
		}
		log.Fatalf("'go mod graph' shows dependencies but vendor folder absent after a 'go mod vendor'")
	}
	run(exec.Command("git", "add", "vendor"))
	msg := fmt.Sprintf("go mod vendor from a go.mod with SHA-256 of %v", modSum)
	run(exec.Command("git", "commit", "-m", msg))
	run(exec.Command("git", "tag", "-a", ref, "-m", msg))
	run(exec.Command("git", "push", "origin", ref))
}

func run(cmd *exec.Cmd) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Running %v: %v, %s", cmd.Args, err, out)
	}
}