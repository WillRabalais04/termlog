package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gen "github.com/WillRabalais04/terminalLog/api/gen"
	grpc_adapter "github.com/WillRabalais04/terminalLog/internal/adapters/grpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc_adapter.NewClient(ctx, "localhost:9090")

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := gen.NewLogServiceClient(conn)

	reader := bufio.NewReader(os.Stdin)
	var text string
	var response *gen.LogResponse

	for {
		fmt.Print("Enter text: ")
		text, _ = reader.ReadString('\n')
		if text == "" || text == "exit" {
			break
		}
		request := &gen.LogEntry{
			Command:          text,
			ExitCode:         0,
			Timestamp:        time.Now().Unix(),
			User:             "testuser",
			WorkingDirectory: "/home/testuser",
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		response, err = client.Log(ctx, request)
		cancel()

		if err != nil {
			log.Printf("could not log: %v", err)
		} else {
			log.Printf("Server Response: Success=%t", response.GetSuccess())
		}
	}

}
