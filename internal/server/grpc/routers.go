package grpcapi

import (
	"context"
	"fmt"
	api "github.com/ennwy/calendar/internal/server"
	pb "github.com/ennwy/calendar/internal/server/grpc/google"
	"github.com/ennwy/calendar/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

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

func (s *GRPCServer) ListEvents(ctx context.Context, event *pb.Event) (*pb.Events, error) {
	events, err := s.App.ListUserEvents(ctx, event.Owner.Name)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	result := &pb.Events{
		Events: make([]*pb.Event, 0, len(events)),
	}
	l.Info("router: list events: events len:", len(events))
	for _, e := range events {
		pbEvent := &pb.Event{
			ID: e.ID,
			Owner: &pb.User{
				ID:   e.Owner.ID,
				Name: e.Owner.Name,
			},
			Title:  e.Title,
			Start:  timestamppb.New(e.Start),
			Finish: timestamppb.New(e.Finish),
			Notify: e.Notify,
		}
		result.Events = append(result.Events, pbEvent)
	}

	return result, nil
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
	start := e.GetStart().AsTime().In(time.UTC)
	finish := e.GetFinish().AsTime().In(time.UTC)

	if !start.Before(finish) {
		return nil, api.ErrTime
	}

	return &storage.Event{
		ID: e.GetID(),
		Owner: storage.User{
			ID:   e.GetOwner().GetID(),
			Name: e.GetOwner().GetName(),
		},
		Title:  e.GetTitle(),
		Start:  start,
		Finish: finish,
		Notify: absNotify(e.GetNotify()),
	}, nil
}
