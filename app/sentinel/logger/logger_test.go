package logger

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"

	pb "github.com/vmware-tanzu/secrets-manager/app/sentinel/logger/generated"
	"google.golang.org/grpc"
)

type MockLogServiceServer struct {
	pb.UnimplementedLogServiceServer
	ReceivedMessage string
}

func (s *MockLogServiceServer) SendLog(ctx context.Context, in *pb.LogRequest) (*pb.LogResponse, error) {
	s.ReceivedMessage = in.Message
	return &pb.LogResponse{}, nil
}

func TestCreateLogServer(t *testing.T) {
	server := &MockLogServiceServer{}
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("error creating log server: %v", err)
	}
	defer lis.Close()

	os.Setenv("SENTINEL_LOGGER_URL", lis.Addr().String())

	grpcServer := grpc.NewServer()
	pb.RegisterLogServiceServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("failed to serve: %v", err)
		}
	}()
	defer grpcServer.Stop()

	CreateLogServer()
}

func TestSendLogMessage(t *testing.T) {
	server := &MockLogServiceServer{}
	grpcServer := grpc.NewServer()
	pb.RegisterLogServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("error creating log server: %v", err)
	}
	defer lis.Close()

	os.Setenv("SENTINEL_LOGGER_URL", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("failed to serve: %v", err)
		}
	}()
	defer grpcServer.Stop()

	message := "Test log message"

	ErrorLn(message)
	if !strings.HasPrefix(server.ReceivedMessage, "[SENTINEL_ERROR]") && !strings.Contains(server.ReceivedMessage, message) {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, message)
	}

	FatalLn(message)
	if !strings.HasPrefix(server.ReceivedMessage, "[SENTINEL_FATAL]") && !strings.Contains(server.ReceivedMessage, message) {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, message)
	}

	WarnLn(message)
	if !strings.HasPrefix(server.ReceivedMessage, "[SENTINEL_WARN]") && !strings.Contains(server.ReceivedMessage, message) {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, message)
	}

	InfoLn(message)
	if !strings.HasPrefix(server.ReceivedMessage, "[SENTINEL_INFO]") && !strings.Contains(server.ReceivedMessage, message) {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, message)
	}

	AuditLn(message)
	if !strings.HasPrefix(server.ReceivedMessage, "[SENTINEL_AUDIT]") && !strings.Contains(server.ReceivedMessage, message) {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, message)
	}

	DebugLn(message)
	if !strings.HasPrefix(server.ReceivedMessage, "[SENTINEL_DEBUG]") && !strings.Contains(server.ReceivedMessage, message) {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, message)
	}

	TraceLn(message)
	if !strings.HasPrefix(server.ReceivedMessage, "[SENTINEL_TRACE]") && !strings.Contains(server.ReceivedMessage, message) {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, message)
	}
}

func TestLogServiceSendLog(t *testing.T) {
	// Define a mock LogRequest
	mockRequest := &pb.LogRequest{
		Message: "Test log message",
	}

	// Start a mock gRPC server
	server := &MockLogServiceServer{}
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("error creating log server: %v", err)
	}
	defer lis.Close()

	os.Setenv("SENTINEL_LOGGER_URL", lis.Addr().String())

	grpcServer := grpc.NewServer()
	pb.RegisterLogServiceServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("failed to serve: %v", err)
		}
	}()
	defer grpcServer.Stop()

	// Create a gRPC client
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()
	client := pb.NewLogServiceClient(conn)

	// Call the SendLog method of the LogServiceServer
	response, err := client.SendLog(context.Background(), mockRequest)
	if err != nil {
		t.Fatalf("SendLog failed: %v", err)
	}

	// Check response
	if response == nil {
		t.Error("Response should not be nil")
	}
	if response.Message != "" {
		t.Error("Response message should be empty")
	}
	if server.ReceivedMessage != mockRequest.Message {
		t.Errorf("Received message (%s) does not match sent message (%s)", server.ReceivedMessage, mockRequest.Message)
	}
}
