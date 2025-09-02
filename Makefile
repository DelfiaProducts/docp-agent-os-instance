#uninstall updater linux local
build:
	@echo "Build project"
	@echo "Build manager"
	@go build -o manager cmd/manager/main.go
	@echo "Build agent"
	@go build -o agent cmd/agent/main.go
	@echo "Build success"