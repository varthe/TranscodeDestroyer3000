package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type logMessage struct {
	level string
	text string
	args []any
}

var (
	queue []logMessage
	mu sync.Mutex
	cond = sync.NewCond(&mu)
)

func init() {
	file, err := os.OpenFile("/proxy/logs/fmq.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Fatal("failed to open log file: %v", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))

	go func() {
		for {
			mu.Lock()
			for len(queue) == 0 {
				cond.Wait()
			}
			msg := queue[0]
			queue = queue[1:]
			mu.Unlock()

			log.Printf("[%s] %s\n", msg.level, fmt.Sprintf(msg.text, msg.args...))
		}
	}()
}

func logAsync(level, msg string, args ...any) {
	mu.Lock()
	queue = append(queue, logMessage{
		level: level,
		text: msg,
		args: args,
	})
	cond.Signal()
	mu.Unlock()
}

func Info(msg string, args ...any) {
	logAsync("INFO", msg, args...)
}

func Debug(msg string, args ...any) {
	logAsync("DEBUG", msg, args...)
}

func Fatal(msg string, args ...any) {
	log.Println("[FATAL] " + fmt.Sprintf(msg, args...))
	os.Exit(1)
}