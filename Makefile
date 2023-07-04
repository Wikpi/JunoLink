all:
	mkdir -p ./build
	go build -o ./build ./cmd/main.go 

add:
	go run ./cmd/main.go add -block $(b)
print:
	go run ./cmd/main.go print