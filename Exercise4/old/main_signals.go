package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func primary() {

	// If statements som ikke kjører så lenge den får signal fra primary
	// If file not updated in x s
	// Logikk primary

	// Spawn Backup
	backup_command_process := exec.Command("gnome-terminal", "--", "go", "run", "./main.go")
	backup_command_process.Stdout = os.Stdout
	backup_command_process.Stderr = os.Stderr
	backup_command_process.Start()

	// Counter variable
	var counter int = 0

	for {

		// Hearbeat
		backup_command_process.Process.Signal(syscall.SIGUSR1)

		// Actual work
		fmt.Printf("Counter: %d\n", counter)
		counter++

		<-time.NewTimer(time.Millisecond * 500).C

		if counter == 100 {
			backup_command_process.Process.Kill()
			return
		}
	}
}

func backup() {
	// If statements som ikke kjører så lenge den får signal fra primary
	// If file not updated in x s
	// Samme logikk som primary

	watchdog_timer := time.NewTimer(time.Second)

	notify_channel := make(chan os.Signal, 1)
	signal.Notify(notify_channel, syscall.SIGUSR1)

	for {
		select {
		case sig := <-notify_channel:
			watchdog_timer.Reset(time.Second)
			fmt.Println("Backup reset!: ", sig)
		case <-watchdog_timer.C:
			return
		}
	}
}

func main() {
	// Kanal for teller
	// Kanal for heartbeats fra primary
	// Kanal for heartbeats fra backup

	backup()
	primary()
}
