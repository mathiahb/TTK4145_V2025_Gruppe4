import (
	"fmt"
	"log"
	"os"
)

func primary() {

	// If statements som ikke kjører så lenge den får signal fra primary
	// If file not updated in x s
	// Logikk primary
}

func backup(fileInfo os.FileInfo) {

	modificationTime := fileInfo.ModTime()

	if time.Since(modificationTime) > 2*time.Second {

	}

	// If statements som ikke kjører så lenge den får signal fra primary
	// If file not updated in x s
	// Samme logikk som primary
}

func main() {

	fileInfo, err := os.Stat("pp.txt")

	// Kanal for teller
	// Kanal for heartbeats fra primary
	// Kanal for heartbeats fra backup
	go primary(fileInfo)
	go backup()

}