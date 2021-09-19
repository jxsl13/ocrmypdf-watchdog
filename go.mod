module github.com/jxsl13/ocrmypdf-watchdog

go 1.16

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/google/uuid v1.3.0
	github.com/jxsl13/simple-configo v1.23.0
)

replace (
	github.com/jxsl13/ocrmypdf-watchdog/config => ./config
	github.com/jxsl13/ocrmypdf-watchdog/internal => ./internal
)
