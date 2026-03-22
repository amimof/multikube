// Package badger provides a badger based repository implementation
package badger

import (
	"context"
	"errors"

	"github.com/amimof/multikube/pkg/repository"
	"github.com/dgraph-io/badger/v4"
)

type DB struct {
	db *badger.DB
}

type txn struct{ t *badger.Txn }

// Keys implements [repository.Txn].
func (x txn) Keys(prefix []byte) ([][]byte, error) {
	opts := badger.DefaultIteratorOptions
	it := x.t.NewIterator(opts)
	defer it.Close()
	var out [][]byte
	it.Seek(prefix)
	for it.ValidForPrefix(prefix) {
		key := it.Item().Key()
		out = append(out, key)
		it.Next()
	}
	return out, nil
}

// List implements [repository.Txn].
// TODO: limit does nothing here
func (x txn) List(prefix []byte, limit int32) ([][]byte, error) {
	opts := badger.DefaultIteratorOptions
	it := x.t.NewIterator(opts)
	defer it.Close()

	var out [][]byte
	it.Seek(prefix)
	for it.ValidForPrefix(prefix) {
		item := it.Item()
		err := item.Value(func(val []byte) error {
			out = append(out, val)
			return nil
		})
		if err != nil {
			return out, err
		}
		it.Next()
	}
	return out, nil
}

func (x txn) Get(key []byte) ([]byte, error) {
	it, err := x.t.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	var out []byte
	err = it.Value(func(v []byte) error {
		out = append(out[:0], v...) // COPY value out
		return nil
	})
	return out, err
}

func (x txn) Set(key, val []byte) error {
	return x.t.Set(key, val)
}

func (x txn) Delete(key []byte) error {
	return x.t.Delete(key)
}

// func (x txn) List() ([][]byte, error) {
// 	return x.t.Delete(key)
// }

func (d *DB) Update(ctx context.Context, fn func(repository.Txn) error) error {
	return d.db.Update(func(t *badger.Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return fn(txn{t: t})
	})
}

func (d *DB) View(ctx context.Context, fn func(repository.Txn) error) error {
	return d.db.View(func(t *badger.Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return fn(txn{t: t})
	})
}

func New(db *badger.DB) repository.DB {
	return &DB{
		db: db,
	}
}
