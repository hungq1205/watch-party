# User service
start cmd.exe /K "cd user-service && go run main.go"

# Message service
start cmd.exe /K "cd message-service && npm start"

# Render service
start cmd.exe /K "cd render-service && go run main.go"