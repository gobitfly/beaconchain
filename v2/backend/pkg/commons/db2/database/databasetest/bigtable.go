package databasetest

import (
	"context"
	"testing"

	"cloud.google.com/go/bigtable"
	"cloud.google.com/go/bigtable/bttest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewBigTable(t testing.TB) (*bigtable.Client, *bigtable.AdminClient) {
	t.Helper()

	// use the same value as the emulated cmd
	// https://github.com/googleapis/google-cloud-go/blob/bigtable/v1.31.0/bigtable/cmd/emulator/cbtemulator.go#L39
	const maxMsgSize = 256 * 1024 * 1024 // 256 MiB
	srv, err := bttest.NewServer("localhost:0", grpc.MaxRecvMsgSize(maxMsgSize), grpc.MaxSendMsgSize(maxMsgSize))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	t.Cleanup(func() { srv.Close() })

	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}

	project, instance := "proj", "instance"
	adminClient, err := bigtable.NewAdminClient(ctx, project, instance, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}

	client, err := bigtable.NewClientWithConfig(ctx, project, instance, bigtable.ClientConfig{MetricsProvider: bigtable.NoopMetricsProvider{}}, option.WithGRPCConn(conn), option.WithGRPCConnectionPool(50))
	if err != nil {
		t.Fatal(err)
	}

	return client, adminClient
}
