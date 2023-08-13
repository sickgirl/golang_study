package main

func main() {
	server := Init()

	server.Run(":8080")
}
