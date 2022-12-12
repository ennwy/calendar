package grpcapi

import (
	"context"
	"fmt"
	api "github.com/ennwy/calendar/internal/server"
	pb "github.com/ennwy/calendar/internal/server/grpc/google"
	"github.com/ennwy/calendar/internal/storage"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	day   = time.Hour * 24
	week  = 7 * day
	month = 30 * day
)

// pb.StorageServer embedded in GRPCServer so always implemented, be careful
var _ pb.StorageServer = (*GRPCServer)(nil)

func (s *GRPCServer) CreateEvent(ctx context.Context, e *pb.Event) (*emptypb.Empty, error) {
	event, err := getEvent(e)
	if err != nil {
		return nil, err
	}

	err = s.App.CreateEvent(ctx, event)
	if err != nil {
		l.Error(err)
	}

	return nil, err
}

func (s *GRPCServer) ListEvents(ctx context.Context, u *pb.User) (*pb.Events, error) {
	events, err := s.App.ListUserEvents(ctx, u.Name)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	pbEvents := getPBEvents(events)

	return pbEvents, nil
}

func (s *GRPCServer) ListUpcoming(ctx context.Context, u *pb.Upcoming) (*pb.Events, error) {
	events, err := s.App.ListUsersUpcoming(
		ctx,
		u.GetOwner().GetName(),
		getUntil(u.GetUntil().Number()),
	)

	if err != nil {
		return nil, fmt.Errorf("list user's upcoming events: %w", err)
	}

	pbEvents := getPBEvents(events)

	return pbEvents, err
}

func (s *GRPCServer) UpdateEvent(ctx context.Context, e *pb.Event) (*emptypb.Empty, error) {
	event, err := getEvent(e)
	if err != nil {
		return nil, fmt.Errorf("updating event: %w", err)
	}

	err = s.App.UpdateEvent(ctx, event)
	if err != nil {
		l.Error(err)
	}

	return nil, err
}

func (s *GRPCServer) DeleteEvent(ctx context.Context, e *pb.Event) (*emptypb.Empty, error) {
	err := s.App.DeleteEvent(ctx, e.ID)
	if err != nil {
		l.Error(err)
	}
	return nil, err
}

// SUB FUNCTIONS
func absNotify(n int32) int32 {
	if n > 0 {
		return n
	}
	return -n
}

func getEvent(e *pb.Event) (*storage.Event, error) {
	start := e.GetStart().AsTime().Round(time.Minute)
	finish := time.Unix(e.GetFinish().GetSeconds(), 0).Round(time.Minute)

	if !start.Before(finish) {
		return nil, api.ErrTime
	}

	return &storage.Event{
		ID:     e.GetID(),
		Title:  e.GetTitle(),
		Start:  start,
		Finish: finish,
		Notify: absNotify(e.GetNotify()),
		Owner: storage.User{
			ID:   e.GetOwner().GetID(),
			Name: e.GetOwner().GetName(),
		},
	}, nil
}

func getPBEvents(events []storage.Event) *pb.Events {
	result := &pb.Events{
		Events: make([]*pb.Event, 0, len(events)),
	}

	for _, e := range events {
		pbEvent := getPBEvent(e)
		result.Events = append(result.Events, pbEvent)
	}

	return result
}

func getPBEvent(event storage.Event) *pb.Event {
	return &pb.Event{
		ID:     event.ID,
		Title:  event.Title,
		Start:  timestamppb.New(event.Start),
		Finish: timestamppb.New(event.Finish),
		Notify: event.Notify,
		Owner: &pb.User{
			ID:   event.Owner.ID,
			Name: event.Owner.Name,
		},
	}
}

func getUntil(n protoreflect.EnumNumber) time.Duration {
	switch int32(n) {
	case 0:
		return day
	case 1:
		return week
	case 2:
		return month
	default:
		return day
	}
}
