package grpcapi

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeConn struct {
	lastMethod string
}

func (c *fakeConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	c.lastMethod = method
	switch out := reply.(type) {
	case *NotificationResponse:
		out.NotificationId = "notif"
		out.Status = Status_SENT
	case *ListNotificationsResponse:
		out.Notifications = []*NotificationResponse{{NotificationId: "notif"}}
	default:
	}
	return nil
}

func (c *fakeConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming not supported")
}

func TestNotificationServiceClientCoverage(t *testing.T) {
	t.Helper()
	client := NewNotificationServiceClient(&fakeConn{})
	ctx := context.Background()
	if _, err := client.SendNotification(ctx, &NotificationRequest{}); err != nil {
		t.Fatalf("SendNotification error: %v", err)
	}
	if _, err := client.GetNotificationStatus(ctx, &GetNotificationStatusRequest{}); err != nil {
		t.Fatalf("GetNotificationStatus error: %v", err)
	}
	if _, err := client.ListNotifications(ctx, &ListNotificationsRequest{}); err != nil {
		t.Fatalf("ListNotifications error: %v", err)
	}
	if _, err := client.RescheduleNotification(ctx, &RescheduleNotificationRequest{}); err != nil {
		t.Fatalf("RescheduleNotification error: %v", err)
	}
	if _, err := client.CancelNotification(ctx, &CancelNotificationRequest{}); err != nil {
		t.Fatalf("CancelNotification error: %v", err)
	}
}

type coverageServer struct {
	UnimplementedNotificationServiceServer
	sendCalled       bool
	statusCalled     bool
	listCalled       bool
	rescheduleCalled bool
	cancelCalled     bool
}

func (s *coverageServer) SendNotification(context.Context, *NotificationRequest) (*NotificationResponse, error) {
	s.sendCalled = true
	return &NotificationResponse{NotificationId: "id"}, nil
}

func (s *coverageServer) GetNotificationStatus(context.Context, *GetNotificationStatusRequest) (*NotificationResponse, error) {
	s.statusCalled = true
	return &NotificationResponse{NotificationId: "id"}, nil
}

func (s *coverageServer) ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error) {
	s.listCalled = true
	return &ListNotificationsResponse{}, nil
}

func (s *coverageServer) RescheduleNotification(context.Context, *RescheduleNotificationRequest) (*NotificationResponse, error) {
	s.rescheduleCalled = true
	return &NotificationResponse{NotificationId: "id"}, nil
}

func (s *coverageServer) CancelNotification(context.Context, *CancelNotificationRequest) (*NotificationResponse, error) {
	s.cancelCalled = true
	return &NotificationResponse{NotificationId: "id"}, nil
}

func TestNotificationServiceServerHandlers(t *testing.T) {
	t.Helper()
	server := &coverageServer{}
	ctx := context.Background()

	decoder := func(interface{}) error { return nil }

	if _, err := _NotificationService_SendNotification_Handler(server, ctx, decoder, nil); err != nil {
		t.Fatalf("Send handler error: %v", err)
	}
	if _, err := _NotificationService_GetNotificationStatus_Handler(server, ctx, decoder, nil); err != nil {
		t.Fatalf("Status handler error: %v", err)
	}
	if _, err := _NotificationService_ListNotifications_Handler(server, ctx, decoder, nil); err != nil {
		t.Fatalf("List handler error: %v", err)
	}
	if _, err := _NotificationService_RescheduleNotification_Handler(server, ctx, decoder, nil); err != nil {
		t.Fatalf("Reschedule handler error: %v", err)
	}
	if _, err := _NotificationService_CancelNotification_Handler(server, ctx, decoder, nil); err != nil {
		t.Fatalf("Cancel handler error: %v", err)
	}

	if !(server.sendCalled && server.statusCalled && server.listCalled && server.rescheduleCalled && server.cancelCalled) {
		t.Fatalf("expected all server methods to be called")
	}
}

func TestRegisterNotificationServiceServer(t *testing.T) {
	t.Helper()
	grpcServer := grpc.NewServer()
	RegisterNotificationServiceServer(grpcServer, &coverageServer{})
	grpcServer.Stop()
}

func TestUnimplementedServerResponses(t *testing.T) {
	t.Helper()
	server := UnimplementedNotificationServiceServer{}
	ctx := context.Background()
	assertUnimplemented := func(err error) {
		if err == nil || status.Code(err) != codes.Unimplemented {
			t.Fatalf("expected unimplemented error, got %v", err)
		}
	}
	_, err := server.SendNotification(ctx, &NotificationRequest{})
	assertUnimplemented(err)
	_, err = server.GetNotificationStatus(ctx, &GetNotificationStatusRequest{})
	assertUnimplemented(err)
	_, err = server.ListNotifications(ctx, &ListNotificationsRequest{})
	assertUnimplemented(err)
	_, err = server.RescheduleNotification(ctx, &RescheduleNotificationRequest{})
	assertUnimplemented(err)
	_, err = server.CancelNotification(ctx, &CancelNotificationRequest{})
	assertUnimplemented(err)
	server.mustEmbedUnimplementedNotificationServiceServer()
	server.testEmbeddedByValue()
}
