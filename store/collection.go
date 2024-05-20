package store

import (
	"context"
	"database/sql"
	"errors"
)

type Record map[string]interface{}

type Collection struct {
	table string
	db    dbInterface
}

func (c *Collection) CreateRecord(ctx context.Context, record Record) (*Record, error) {
	query := "INSERT INTO " + c.table + " SET ?"
	res, err := c.db.ExecContext(ctx, query, record)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return c.GetRecord(ctx, id)
}

func (c *Collection) GetRecord(ctx context.Context, id interface{}) (*Record, error) {
	var rec Record
	query := "SELECT * FROM " + c.table + " WHERE id = ?"
	err := c.db.GetContext(ctx, &rec, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundErr(err)
		}

		return nil, err
	}
	return &rec, nil
}

func (c *Collection) UpdateRecord(ctx context.Context, id interface{}, record Record) (*Record, error) {
	query := "UPDATE " + c.table + " SET ? WHERE id = ?"
	_, err := c.db.ExecContext(ctx, query, record, id)
	if err != nil {
		return nil, err
	}
	return c.GetRecord(ctx, id)
}

func (c *Collection) DeleteRecord(ctx context.Context, id interface{}) error {
	query := "DELETE FROM " + c.table + " WHERE id = ?"
	_, err := c.db.ExecContext(ctx, query, id)
	return err
}

func (c *Collection) FindOne(ctx context.Context, filter interface{}) (*Record, error) {
	var rec Record
	query := "SELECT * FROM " + c.table + " WHERE ? LIMIT 1"
	err := c.db.GetContext(ctx, &rec, query, filter)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (c *Collection) Find(ctx context.Context, filter interface{}) ([]Record, error) {
	var recs []Record
	query := "SELECT * FROM " + c.table + " WHERE ?"
	err := c.db.SelectContext(ctx, &recs, query, filter)
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func (c *Collection) Count(ctx context.Context, filter interface{}) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM " + c.table + " WHERE ?"
	err := c.db.GetContext(ctx, &count, query, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Collection) Exists(ctx context.Context, filter interface{}) (bool, error) {
	count, err := c.Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
