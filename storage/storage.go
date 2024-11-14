package storage

import (
	"context"
	"database/sql"
	"errors"
	_ "errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"log/slog"
	"net/url"
)

const moduleName = "storage"

type Storage struct {
	lg *slog.Logger
	db *sql.DB
}

func New(
	lg *slog.Logger,
	username string,
	password string,
	address string,
	database string,
) (*Storage, error) {
	dsn := (&url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(username, password),
		Host:   address,
		Path:   database,
	}).String()

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("init db: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %v", err)
	}

	return &Storage{
		lg: lg.With("module", moduleName),
		db: sqlDB,
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddProductFriend(ctx context.Context, productFriend ProductFriend) error {
	query := `INSERT INTO products (name, hobby, price) VALUES ($1, $2, $3)`
	if _, err := s.db.ExecContext(
		ctx,
		query,
		productFriend.Name,
		productFriend.Hobby,
		productFriend.Price,
	); err != nil {
		return fmt.Errorf("add product friend: %v", err)
	}
	return nil
}

func (s *Storage) MigriteUP() (int, error) {
	migrations := &migrate.FileMigrationSource{
		Dir: "dirMigrite",
	}
	n, err := migrate.Exec(s.db, "postgres", migrations, migrate.Up)
	if err != nil {
		s.lg.Error("ошибка", err)
		return n, err
	}
	return n, nil
}

func (s *Storage) MigriteDOWN() (int, error) {
	migrations := &migrate.FileMigrationSource{
		Dir: "dirMigrite",
	}
	n, err := migrate.Exec(s.db, "postgres", migrations, migrate.Down)
	if err != nil {
		s.lg.Error("ошибка", err)
		return n, err
	}
	return n, nil
}

func (s *Storage) GetQuery(query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(context.Background(), query, args...)
}

type Product struct {
	ID    int
	Name  string
	Hobby string
	Price int
}

func (s *Storage) GetZZZ() ([]Product, error) {

	rows, err := s.GetQuery(`SELECT * FROM products`)
	if err != nil {
		slog.Error("Failed to execute query in GetZZZ", err)
		return nil, err
	}

	defer rows.Close()

	var product []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Hobby, &p.Price); err != nil {
			slog.Error("Failed to scan row in GetZZZ", err)
			return nil, err
		}
		product = append(product, p)
	}
	if err := rows.Err(); err != nil {
		slog.Error("Error iterating rows in GetZZZ", err)
		return nil, err
	}
	return product, nil
}

func (s *Storage) Getdelete(id int) error {
	result, err := s.db.ExecContext(context.Background(), `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		slog.Error("err")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("err")
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
