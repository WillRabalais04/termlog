package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gen "github.com/WillRabalais04/terminalLog/api/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		if text == "" {
			break
		}
		request := &gen.LogEntry{
			Command:          text,
			ExitCode:         0,
			Timestamp:        time.Now().Unix(),
			User:             "testuser",
			WorkingDirectory: "/home/testuser",
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err = client.Log(ctx, request)
		if err != nil {
			log.Fatalf("could not log: %v", err)
		}
		log.Printf("Server Response: Success=%t", response.GetSuccess())
	}

}
