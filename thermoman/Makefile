.PHONY : clean build run

clean:
	go clean

build:
	GOOS=darwin GOARCH=amd64 go build -o bin/ThermoMan-darwin-amd64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/ThermoMan-linux-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/ThermoMan-windows-amd64 main.go

run:
	go run main.go
