package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/lib/pq"
	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
)

type DB struct {
	conn *sql.DB
}

func NewDB(ctx context.Context) (*DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "boutique_db")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) GetAllProducts(ctx context.Context) ([]*pb.Product, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, name, description, picture_url, price_usd_currency_code, 
                price_usd_units, price_usd_nanos, categories 
         FROM products ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*pb.Product
	for rows.Next() {
		p := &pb.Product{}
		var categories []string

		err := rows.Scan(
			&p.Id,
			&p.Name,
			&p.Description,
			&p.Picture,
			&p.PriceUsd.CurrencyCode,
			&p.PriceUsd.Units,
			&p.PriceUsd.Nanos,
			pq.Array(&categories),
		)
		if err != nil {
			return nil, err
		}

		p.Categories = categories
		products = append(products, p)
	}

	return products, rows.Err()
}

func (d *DB) GetProduct(ctx context.Context, id string) (*pb.Product, error) {
	row := d.conn.QueryRowContext(ctx,
		`SELECT id, name, description, picture_url, price_usd_currency_code,
                price_usd_units, price_usd_nanos, categories
         FROM products WHERE id = $1`, id)

	p := &pb.Product{}
	var categories []string

	err := row.Scan(
		&p.Id,
		&p.Name,
		&p.Description,
		&p.Picture,
		&p.PriceUsd.CurrencyCode,
		&p.PriceUsd.Units,
		&p.PriceUsd.Nanos,
		pq.Array(&categories),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", id)
		}
		return nil, err
	}

	p.Categories = categories
	return p, nil
}

func (d *DB) SearchProducts(ctx context.Context, query string) ([]*pb.Product, error) {
	searchTerm := "%" + query + "%"
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, name, description, picture_url, price_usd_currency_code,
                price_usd_units, price_usd_nanos, categories
         FROM products
         WHERE LOWER(name) LIKE LOWER($1) OR LOWER(description) LIKE LOWER($1)
         ORDER BY name`, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*pb.Product
	for rows.Next() {
		p := &pb.Product{}
		var categories []string

		err := rows.Scan(
			&p.Id,
			&p.Name,
			&p.Description,
			&p.Picture,
			&p.PriceUsd.CurrencyCode,
			&p.PriceUsd.Units,
			&p.PriceUsd.Nanos,
			pq.Array(&categories),
		)
		if err != nil {
			return nil, err
		}

		p.Categories = categories
		products = append(products, p)
	}

	return products, rows.Err()
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
