package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func main() {
	cmd := exec.Command("./matrix", "-lines", "20", "-speed", "0.05")
	cmd.Env = append(os.Environ(), "TERM=xterm")
	f, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}

	byteCount := 0
	buf := make([]byte, 32*1024)
	
	go func() {
		for {
			n, err := f.Read(buf)
			if n > 0 {
				byteCount += n
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				break
			}
		}
	}()

	cmd.Wait()
	log.Printf("Total bytes written for 20 ticks: %d\n", byteCount)
}
