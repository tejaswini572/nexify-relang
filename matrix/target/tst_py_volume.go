package main

import (
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("python3", "../source/rain.py")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	byteCount := 0
	buf := make([]byte, 32*1024)
	
	// Read for exactly 20 ticks. Since speed is 0.1, we can just read for 2.1 seconds.
	// Actually, just reading 20 lines is better since python prints 1 line per tick.
	lineCount := 0
	for lineCount < 20 {
		n, err := stdout.Read(buf)
		if n > 0 {
			byteCount += n
			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					lineCount++
					if lineCount == 20 {
						break
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	cmd.Process.Kill()
	log.Printf("Python: Total bytes written for 20 ticks: %d\n", byteCount)
}
