package utils

import "fmt"

func FormatSize(size int64) string {
	switch {
	case size > 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
	case size > 1024:
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}
