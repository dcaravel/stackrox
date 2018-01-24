package service

import (
	"bitbucket.org/stack-rox/apollo/central/db"
	"bitbucket.org/stack-rox/apollo/central/notifications"
	"bitbucket.org/stack-rox/apollo/generated/api/v1"
	"bitbucket.org/stack-rox/apollo/pkg/grpc/auth"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hashicorp/golang-lru"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const cacheSize = 100

// NewBenchmarkResultsService returns the BenchmarkResultsService API.
func NewBenchmarkResultsService(storage db.Storage, notificationsProcessor *notifications.Processor) *BenchmarkResultsService {
	cache, err := lru.New(cacheSize)
	if err != nil {
		// This only happens in extreme cases (at this time, for invalid size only).
		panic(err)
	}
	return &BenchmarkResultsService{
		resultStore:   storage,
		scheduleStore: storage,
		cache:         cache,
		notificationsProcessor: notificationsProcessor,
	}
}

// BenchmarkResultsService is the struct that manages the benchmark results API
type BenchmarkResultsService struct {
	resultStore            db.BenchmarkScansStorage
	scheduleStore          db.BenchmarkScheduleStorage
	notificationsProcessor *notifications.Processor

	cache *lru.Cache
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *BenchmarkResultsService) RegisterServiceServer(grpcServer *grpc.Server) {
	v1.RegisterBenchmarkResultsServiceServer(grpcServer, s)
}

// RegisterServiceHandlerFromEndpoint registers this service with the given gRPC Gateway endpoint.
func (s *BenchmarkResultsService) RegisterServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return v1.RegisterBenchmarkResultsServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

// PostBenchmarkResult inserts a new benchmark result into the system
func (s *BenchmarkResultsService) PostBenchmarkResult(ctx context.Context, request *v1.BenchmarkResult) (*empty.Empty, error) {
	identity, err := auth.FromContext(ctx)
	if err != nil || identity.IdentityType.ServiceType != v1.ServiceType_SENSOR_SERVICE {
		return nil, status.Error(codes.Unauthenticated, "only sensors are allowed")
	}
	if err := s.resultStore.AddBenchmarkResult(request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if request.GetReason() == v1.BenchmarkReason_SCHEDULED {
		if _, ok := s.cache.Get(request.GetScanId()); ok {
			// This means that the scan id has already been processed and an alert about benchmarks coming in was already sent
			return &empty.Empty{}, nil
		}
		s.cache.Add(request.GetScanId(), struct{}{})
		schedule, exists, err := s.scheduleStore.GetBenchmarkSchedule(request.GetName())
		if err != nil {
			log.Errorf("Error retrieving benchmark schedule %v: %+v", request.GetName(), err)
			return &empty.Empty{}, nil
		} else if !exists {
			log.Errorf("Benchmark schedule %v does not exist", request.GetName())
			return &empty.Empty{}, nil
		}
		s.notificationsProcessor.ProcessBenchmark(schedule)
	}
	return &empty.Empty{}, nil
}
