package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const startupAnimationPath = "apps/lib/generate/frames"

func main() {
	runAnimation()
}

func runAnimation() error {
	basePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	frames, err := os.ReadDir(filepath.Join(basePath, startupAnimationPath))
	if err != nil {
		return fmt.Errorf("failed to read frames: %w", err)
	}

	for _, frame := range frames {
		frameContent, err := os.ReadFile(filepath.Join(basePath, startupAnimationPath, frame.Name()))
		if err != nil {
			return fmt.Errorf("failed to read frame %s: %w", frame.Name(), err)
		}

		// Clear the screen
		fmt.Print("\033[H\033[2J")

		// Print the frame
		fmt.Println(string(frameContent))
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Print("\033[H\033[2J")

	return nil
}
