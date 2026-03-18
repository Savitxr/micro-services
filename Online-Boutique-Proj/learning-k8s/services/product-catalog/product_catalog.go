package main

import (
	"context"
	"time"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type productCatalog struct {
	pb.UnimplementedProductCatalogServiceServer
	db *DB
}

func (p *productCatalog) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (p *productCatalog) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (p *productCatalog) ListProducts(ctx context.Context, _ *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(extraLatency)

	products, err := p.db.GetAllProducts(ctx)
	if err != nil {
		log.Errorf("failed to get products: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get products")
	}

	return &pb.ListProductsResponse{Products: products}, nil
}

func (p *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(extraLatency)

	product, err := p.db.GetProduct(ctx, req.Id)
	if err != nil {
		log.Errorf("failed to get product %s: %v", req.Id, err)
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
	}
	return product, nil
}

func (p *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(extraLatency)

	products, err := p.db.SearchProducts(ctx, req.Query)
	if err != nil {
		log.Errorf("failed to search products: %v", err)
		return nil, status.Errorf(codes.Internal, "search failed")
	}

	return &pb.SearchProductsResponse{Results: products}, nil
}
