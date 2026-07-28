package main

import (
	"log"
	"os/exec"
	"time"
	"github.com/creack/pty"
)

func main() {
	cmd := exec.Command("./matrix", "-lines", "20", "-speed", "0.05")
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		log.Fatal(err)
	}

	// Wait a bit, then resize
	time.Sleep(200 * time.Millisecond)
	pty.Setsize(ptmx, &pty.Winsize{Cols: 150, Rows: 24})
	
	time.Sleep(200 * time.Millisecond)
	pty.Setsize(ptmx, &pty.Winsize{Cols: 40, Rows: 24})
	
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("Process exited with error: %v", err)
	}
	log.Println("Process finished successfully without crashing during resize.")
}
