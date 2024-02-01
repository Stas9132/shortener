# cmd/shortener

В данной директории будет содержаться код, который скомпилируется в бинарное приложение

# Windows systems
go build -ldflags="-X main.buildVersion=v1.0.0 -X main.buildDate=29.01.2024 -X main.buildCommit=commit1" .\cmd\shortener\main.go

# Linux systems
go build -ldflags="-X main.buildVersion=v1.0.0 -X main.buildDate=29.01.2024 -X main.buildCommit=commit1" ./cmd/shortener/main.go