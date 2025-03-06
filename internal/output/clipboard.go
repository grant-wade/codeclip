package output

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
)

// CopyToTarget copies the formatted content to the specified target
func CopyToTarget(content string, target string) error {
	switch target {
	case "clipboard":
		return copyToClipboard(content)
	case "stdout":
		fmt.Print(content)
		return nil
	default:
		// Assume target is a file path
		return os.WriteFile(target, []byte(content), 0644)
	}
}

// copyToClipboard attempts to copy content to clipboard, falling back to stdout if unavailable
func copyToClipboard(content string) error {
	if !clipboard.Unsupported {
		err := clipboard.WriteAll(content)
		if err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		return nil
	}

	// Clipboard unsupported
	fmt.Fprintln(os.Stderr, "Clipboard not supported on this system. Outputting to stdout:")
	fmt.Print(content)
	return nil
}
