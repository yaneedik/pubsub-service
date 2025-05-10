package service

import (
	"context"

	"github.com/yaneedik/pubsub-service/internal/subpub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PubSubService struct {
	UnimplementedPubSubServer
	bus subpub.SubPub
}

func NewPubSubService(bus subpub.SubPub) *PubSubService {
	return &PubSubService{bus: bus}
}

func (s *PubSubService) Subscribe(req *SubscribeRequest, stream PubSub_SubscribeServer) error {
	sub, err := s.bus.Subscribe(req.Key, func(msg interface{}) {
		if data, ok := msg.(string); ok {
			stream.Send(&Event{Data: data})
		}
	})
	if err != nil {
		return status.Errorf(codes.Internal, "subscribe error: %v", err)
	}
	defer sub.Unsubscribe()

	<-stream.Context().Done()
	return nil
}

func (s *PubSubService) Publish(ctx context.Context, req *PublishRequest) (*emptypb.Empty, error) {
	if err := s.bus.Publish(req.Key, req.Data); err != nil {
		return nil, status.Errorf(codes.Internal, "publish error: %v", err)
	}
	return &emptypb.Empty{}, nil
}