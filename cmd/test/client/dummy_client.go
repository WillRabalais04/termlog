package client_test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/WillRabalais04/terminalLog/api/gen"
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

	client := pb.NewLogServiceClient(conn)

	reader := bufio.NewReader(os.Stdin)
	var text string
	var response *pb.LogResponse

	for {
		fmt.Print("Enter text: ")
		text, _ = reader.ReadString('\n')
		if text == "" || text == "exit" {
			break
		}
		entry := &pb.LogEntry{
			Command:              text,
			ExitCode:             0,
			Timestamp:            time.Now().Unix(),
			User:                 "dummy_client",
			WorkingDirectory:     "/home/testuser",
			Shell_PID:            0,
			ShellUptime:          69,
			PrevWorkingDirectory: "/home/prev",
			EUID:                 0,
			Term:                 "kitty",
			Hostname:             "a",
			SSHClient:            "b",
			TTY:                  "tty01",
			GitRepo:            false,
			GitRepoRoot:          "asdf",
			GitBranch:            "main",
			GitCommit:            "2342aga",
			GitStatus:            "no diff",
			LoggedSuccessfully:   true,
		}
		request := &pb.LogRequest{
			Entry: entry,
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
