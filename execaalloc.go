package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var command string

func StartProfilerService(port int) {
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()
}

func main() {
	slow := flag.Bool("slow", false, "Start slow growing test.")
	flag.Parse()

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		dir = strings.Replace(dir, "\\", "\\\\", -1)
	}
	fmt.Println(dir)

	command = fmt.Sprintf("import sys; sys.path.append('%s'); import handler; print handler.move();", dir)
	fmt.Println(command)

	debug.SetGCPercent(25)

	StartProfilerService(6060)

	if *slow {
		slowerLeak(&command)
		return
	}
	fastLeak(&command)
}

func fastLeak(command *string) {
	for {
		cmd := exec.Command("python", "-c", *command)
		out, err := cmd.Output()
		if err != nil {
			log.Printf("Command '%s' returned error %s.", command, err)
		}
		log.Printf("Output: %s.", string(out))
		time.Sleep(time.Second)
	}
}

func slowerLeak(command *string) {
	// Create command buffers
	var buffers execBufferCache
	// Get the environment
	environment := append(os.Environ())
	// Script executable path
	scriptPath := "python"
	if filepath.Base(scriptPath) == "python" {
		if lp, err := exec.LookPath("python"); err != nil {
			log.Printf("Unable to find executable with name '%s'; Error : '%s'.", "python", err)
			return
		} else {
			scriptPath = lp
		}
	}

	for {
		buffers.Reset()
		cmd := exec.Command(scriptPath, "-c", *command)
		buffers.SetBuffers(cmd)
		cmd.Env = environment

		if err := cmd.Run(); err != nil {
			log.Printf("Error executing command %s.", err)
		}
		// Get the result
		log.Printf("Output: %s.", string(buffers.CmdOutBuf.Bytes()))
		time.Sleep(time.Second)
	}
}

// Contains buffers used by exec
// Each subsequent call will reuse these buffers.
// preventing system from constantly allocating new buffers
type execBufferCache struct {
	CmdInBuf  bytes.Buffer
	CmdOutBuf bytes.Buffer
	CmdErrBuf bytes.Buffer
}

// Reset the buffers to the beginning
// This should be called when the caller is done with the given exec call
func (c *execBufferCache) Reset() {
	c.CmdInBuf.Reset()
	c.CmdOutBuf.Reset()
	c.CmdErrBuf.Reset()
}

// Initialise the command structure with appropriate buffers
// // cmd - exec command structure
func (c *execBufferCache) SetBuffers(cmd *exec.Cmd) {
	cmd.Stdin = &c.CmdInBuf
	cmd.Stdout = &c.CmdOutBuf
	cmd.Stderr = &c.CmdErrBuf
}
