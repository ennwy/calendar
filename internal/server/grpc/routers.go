package grpcapi

import (
	"context"
	pb "github.com/ennwy/calendar/internal/server/grpc/google"
	"github.com/ennwy/calendar/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func (s *GRPCServer) CreateEvent(ctx context.Context, e *pb.Event) (*emptypb.Empty, error) {
	event := &storage.Event{
		OwnerID: e.GetOwnerID(),
		Title:   e.GetTitle(),
		Start:   e.GetStart().AsTime().In(time.UTC),
		Finish:  e.GetFinish().AsTime().In(time.UTC),
	}
	err := s.App.CreateEvent(ctx, event)
	l.Info(event)

	if err != nil {
		l.Error(err)
	}

	return nil, err
}

func (s *GRPCServer) ListEvents(ctx context.Context, event *pb.Event) (*pb.Events, error) {
	t := timestamppb.Now()
	l.Info(t)

	events, err := s.App.ListEvents(ctx, event.OwnerID)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	pbEvents := make([]*pb.Event, 0, len(events))

	var e storage.Event
	for i := range events {
		e = events[i]

		pbEvent := &pb.Event{
			ID:      e.ID,
			OwnerID: e.OwnerID,
			Title:   e.Title,
			Start:   timestamppb.New(e.Start),
			Finish:  timestamppb.New(e.Finish),
		}

		pbEvents = append(pbEvents, pbEvent)
	}

	result := &pb.Events{
		Events: pbEvents,
	}

	return result, nil
}

func (s *GRPCServer) UpdateEvent(ctx context.Context, e *pb.Event) (*emptypb.Empty, error) {
	event := storage.Event{
		ID:     e.GetID(),
		Title:  e.GetTitle(),
		Start:  e.GetStart().AsTime().In(time.UTC),
		Finish: e.GetFinish().AsTime().In(time.UTC),
	}
	err := s.App.UpdateEvent(ctx, event)
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
