package adding

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/kkereziev/notifier/internal/entity"
	pb "github.com/kkereziev/notifier/internal/proto"
	"github.com/kkereziev/notifier/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// IdempotencyHeader is the name of the header that contains the idempotency key.
const IdempotencyHeader = "X-Idempotency-Key"

// Service handles adding notifications logic for the application.
type Service struct {
	logger           *zap.SugaredLogger
	tx               storage.Transactioner
	idempotencyRepo  IdempotencyStorage
	notificationRepo NotificationStorage
	*pb.UnimplementedNotificationServiceServer
}

// NewService is a constructor function for Service.
func NewService(
	logger *zap.SugaredLogger,
	tx storage.Transactioner,
	idempotencyRepo IdempotencyStorage,
	notificationRepo NotificationStorage,
) *Service {
	return &Service{
		logger:           logger,
		tx:               tx,
		idempotencyRepo:  idempotencyRepo,
		notificationRepo: notificationRepo,
	}
}

// SendSMSNotification handles the logic for adding SMS notifications.
func (s *Service) SendSMSNotification(
	ctx context.Context,
	req *pb.SendSMSNotificationRequest,
) (*pb.SendSMSNotificationResponse, error) {
	sms, err := entity.NewSMSNotification(req.GetMessage(), req.GetSendTo())
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	data, err := json.Marshal(sms)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	ent := &entity.Notification{Type: entity.SMSNotificationType, Status: entity.PendingNotificationStatus, Data: data}

	if err := s.handleRequest(ctx, ent); err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.SendSMSNotificationResponse{}, nil
}

// SendSlackNotification handles the logic for adding Slack notifications.
func (s *Service) SendSlackNotification(
	ctx context.Context,
	req *pb.SendSlackNotificationRequest,
) (*pb.SendSlackNotificationResponse, error) {
	slack, err := entity.NewSlackNotification(req.GetMessage())
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	data, err := json.Marshal(slack)
	if err != nil {
		return nil, err
	}

	ent := &entity.Notification{Type: entity.SlackNotificationType, Status: entity.PendingNotificationStatus, Data: data}
	if err := s.handleRequest(ctx, ent); err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.SendSlackNotificationResponse{}, nil
}

// SendMailNotification handles the logic for adding mail notifications.
func (s *Service) SendMailNotification(
	ctx context.Context,
	req *pb.SendMailNotificationRequest,
) (*pb.SendMailNotificationResponse, error) {
	mail, err := entity.NewMailNotification(req.GetMessage(), req.GetSendTo(), req.GetSubject())
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	data, err := json.Marshal(mail)
	if err != nil {
		return nil, err
	}

	ent := &entity.Notification{Type: entity.EmailNotificationType, Status: entity.PendingNotificationStatus, Data: data}
	if err := s.handleRequest(ctx, ent); err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.SendMailNotificationResponse{}, nil
}

func (s *Service) handleRequest(ctx context.Context, notification *entity.Notification) error {
	key, err := extractIdempotencyHeader(ctx)
	if err != nil {
		return err
	}

	s.logger.Infof("Key %s", key)

	return s.tx.Tx(ctx, func(ctx context.Context) error {
		id := uuid.MustParse(key)

		idempotencyKey, err := s.idempotencyRepo.GetByID(ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if idempotencyKey != nil {
			return nil
		}

		if err := s.notificationRepo.InsertOne(ctx, notification); err != nil {
			return err
		}

		return s.idempotencyRepo.InsertOne(ctx, &entity.IdempotencyKey{ID: id})
	})
}

func extractIdempotencyHeader(ctx context.Context) (string, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("request should contain idempotency key")
	}

	keys := headers.Get(IdempotencyHeader)

	if len(keys) == 0 {
		return "", errors.New("request should contain idempotency key")
	}

	return keys[0], nil
}
