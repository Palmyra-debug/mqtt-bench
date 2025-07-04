package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			var color int
			switch i.(string) {
			case "debug":
				color = 36 // голубой
			case "info":
				color = 32 // зелёный
			case "warn":
				color = 33 // жёлтый
			case "error":
				color = 31 // красный
			default:
				color = 0 // без цвета
			}
			return fmt.Sprintf("- \x1b[%dm%-5s\x1b[0m-", color, strings.ToUpper(i.(string)))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
	}
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}

	lvl, err := zerolog.ParseLevel(strings.ToLower(levelStr))
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	Log = zerolog.New(output).
		Level(lvl).
		With().
		Timestamp().
		Logger()

	Log.Info().Msg("Logger initialized")
}
