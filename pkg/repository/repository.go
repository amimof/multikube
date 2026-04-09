// Package repository provides interfaces for implementing storage solutions for types
package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/protoutils"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certv1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
)

var (
	ErrNotFound  = errors.New("item not found")
	ErrIdxExists = errors.New("index already exists")
)

type Resource interface {
	proto.Message
	GetMeta() *metav1.Meta
}

type Txn interface {
	Get(key []byte) ([]byte, error)       // returns a COPY of value
	List([]byte, int32) ([][]byte, error) // returns a COPY of value
	Set(key, val []byte) error
	Delete(key []byte) error
	Keys([]byte) ([][]byte, error)
}

type DB interface {
	View(ctx context.Context, fn func(txn Txn) error) error
	Update(ctx context.Context, fn func(txn Txn) error) error
}

var BackendCodec = ProtoCodec[*backendv1.Backend]{
	New: func() *backendv1.Backend { return &backendv1.Backend{} },
}

func NewBackendRepo[T *backendv1.Backend](db DB) *Repo[*backendv1.Backend] {
	return NewRepo(db, BackendCodec, []byte("backend/"), []byte("i/backend/"), []byte("i/idx/backend"))
}

var CertificateAuthorityCodec = ProtoCodec[*cav1.CertificateAuthority]{
	New: func() *cav1.CertificateAuthority { return &cav1.CertificateAuthority{} },
}

func NewCertificateAuthorityRepo[T *cav1.CertificateAuthority](db DB) *Repo[*cav1.CertificateAuthority] {
	return NewRepo(db, CertificateAuthorityCodec, []byte("certificateauthority/"), []byte("i/certificateauthority/"), []byte("i/idx/certificateauthority"))
}

var CertificateCodec = ProtoCodec[*certv1.Certificate]{
	New: func() *certv1.Certificate { return &certv1.Certificate{} },
}

func NewCertificateRepo[T *certv1.Certificate](db DB) *Repo[*certv1.Certificate] {
	return NewRepo(db, CertificateCodec, []byte("certificate/"), []byte("i/certificate/"), []byte("i/idx/certificate"))
}

var RouteCodec = ProtoCodec[*routev1.Route]{
	New: func() *routev1.Route { return &routev1.Route{} },
}

func NewRouteRepo[T *routev1.Route](db DB) *Repo[*routev1.Route] {
	return NewRepo(db, RouteCodec, []byte("route/"), []byte("i/route/"), []byte("i/idx/route"))
}

var PolicyCodec = ProtoCodec[*policyv1.Policy]{
	New: func() *policyv1.Policy { return &policyv1.Policy{} },
}

func NewPolicyRepo[T *policyv1.Policy](db DB) *Repo[*policyv1.Policy] {
	return NewRepo(db, PolicyCodec, []byte("policy/"), []byte("i/policy/"), []byte("i/idx/policy"))
}

var CredentialCodec = ProtoCodec[*credentialv1.Credential]{
	New: func() *credentialv1.Credential { return &credentialv1.Credential{} },
}

func NewCredentialRepo[T *credentialv1.Credential](db DB) *Repo[*credentialv1.Credential] {
	return NewRepo(db, CredentialCodec, []byte("credential/"), []byte("i/credential/"), []byte("i/idx/credential"))
}

type Codec[T proto.Message] interface {
	Decode([]byte) (T, error)
}

type ProtoCodec[T proto.Message] struct {
	New func() T
}

func (c ProtoCodec[T]) Decode(b []byte) (T, error) {
	msg := c.New()
	if err := proto.Unmarshal(b, msg); err != nil {
		var zero T
		return zero, err
	}
	return msg, nil
}

type Repo[T Resource] struct {
	// db      *badger.DB
	db        DB
	prefix    []byte
	iprefix   []byte
	idxprefix []byte
	Codec     Codec[T]
}

// func NewRepo[T proto.Message](db *badger.DB, codec Codec[T], prefix, iprefix []byte) *Repo[T] {
func NewRepo[T Resource](db DB, codec Codec[T], prefix, iprefix, idxprefix []byte) *Repo[T] {
	return &Repo[T]{
		db:        db,
		prefix:    prefix,
		iprefix:   iprefix,
		idxprefix: idxprefix,
		Codec:     codec,
	}
}

func (r Repo[T]) List(ctx context.Context, limit int32) ([]T, error) {
	var out []T

	err := r.db.View(ctx, func(txn Txn) error {
		kvs, err := txn.List(r.prefix, limit)
		if err != nil {
			return err
		}

		out = make([]T, 0, len(kvs))
		for _, kvp := range kvs {
			obj, derr := r.Codec.Decode(kvp)
			if derr != nil {
				return derr
			}
			out = append(out, obj)
		}
		return nil
	})

	// If out is empty, nextCursor will be nil.
	return out, err
}

func (r Repo[T]) Get(ctx context.Context, id keys.ID) (T, error) {
	var res T

	err := r.db.View(ctx, func(txn Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch id.Tag() {
		case keys.TagUID:
			t, err := r.getByUID(ctx, id)
			if err != nil {
				return err
			}
			res = t

		case keys.TagName:
			t, err := r.getByName(ctx, id)
			if err != nil {
				return err
			}
			res = t
		case keys.TagIdx:
			t, err := r.getByIndex(ctx, id)
			if err != nil {
				return err
			}
			res = t
		default:
			return fmt.Errorf("unsupported  sssid tag: %v", id.Tag())
		}
		return nil
	})
	return res, err
}

func (r *Repo[T]) getByUID(ctx context.Context, id keys.ID) (T, error) {
	key := id.EncodePrefixed(r.prefix)

	var res T
	err := r.db.View(ctx, func(txn Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			}
			return err
		}
		decoded, err := r.Codec.Decode(item)
		if err != nil {
			return err
		}
		res = decoded
		return nil
	})
	if err != nil {
		return res, err
	}
	return res, nil
}

func (r Repo[T]) getByName(ctx context.Context, id keys.ID) (T, error) {
	name := id.EncodePrefixed(r.iprefix)
	var res T

	err := r.db.View(ctx, func(txn Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		idxItem, err := txn.Get(name)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			}
			return err
		}

		var uid keys.ID
		uid, err = keys.Decode(idxItem)
		if err != nil {
			return err
		}

		item, err := txn.Get(uid.EncodePrefixed(r.prefix))
		if err != nil {
			return err
		}
		decoded, err := r.Codec.Decode(item)
		if err != nil {
			return err
		}
		res = decoded
		return nil
	})

	return res, err
}

func (r Repo[T]) getByIndex(ctx context.Context, id keys.ID) (T, error) {
	idx := id.EncodePrefixed(r.idxprefix)
	var res T

	err := r.db.View(ctx, func(txn Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		idxItem, err := txn.Get(idx)
		if err != nil {
			return err
		}

		var uid keys.ID
		uid, err = keys.Decode(idxItem)
		if err != nil {
			return err
		}

		item, err := txn.Get(uid.EncodePrefixed(r.prefix))
		if err != nil {
			return err
		}
		decoded, err := r.Codec.Decode(item)
		if err != nil {
			return err
		}
		res = decoded
		return nil
	})

	return res, err
}

func (r *Repo[T]) Create(ctx context.Context, resource T) (T, error) {
	var res T

	u := uuid.New()
	resource.GetMeta().Uid = u.String()
	resource.GetMeta().Created = timestamppb.Now()
	resource.GetMeta().Updated = timestamppb.Now()
	resource.GetMeta().ResourceVersion = 1

	uid, err := keys.UUID(u)
	if err != nil {
		return res, err
	}

	name, err := keys.Name(resource.GetMeta().GetName())
	if err != nil {
		return res, err
	}

	idx, err := keys.Index(resource.GetMeta().GetName())
	if err != nil {
		return res, err
	}

	err = r.db.Update(ctx, func(txn Txn) error {
		existing, err := r.Get(ctx, name)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return err
			}
		}

		if err == nil {
			return ErrIdxExists
		}

		equal, err := protoutils.SpecEqual(existing, resource)
		if err != nil {
			return err
		}

		protoutils.EnsureMessageField(resource, "status")

		if !equal {
			resource.GetMeta().Generation++
		}

		b, err := proto.Marshal(resource)
		if err != nil {
			return err
		}

		if err := txn.Set(uid.EncodePrefixed(r.prefix), b); err != nil {
			return err
		}

		if err := txn.Set(idx.EncodePrefixed(r.idxprefix), uid.Encode()); err != nil {
			return err
		}

		if err := txn.Set(name.EncodePrefixed(r.iprefix), uid.Encode()); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return res, err
	}

	res, err = r.getByUID(ctx, uid)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r Repo[T]) Delete(ctx context.Context, id keys.ID) error {
	err := r.db.Update(ctx, func(txn Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch id.Tag() {
		case keys.TagUID:
			// Find and delete the index entry that points to this UID
			indexKeys, err := txn.Keys(r.iprefix)
			if err != nil {
				return err
			}

			// Iterate through all index keys to find which one points to this UID
			for _, indexKey := range indexKeys {
				// Read the value stored in this index key (it contains the UID)
				indexValue, err := txn.Get(indexKey)
				if err != nil {
					continue // Skip if we can't read this index key
				}

				// Decode the UID stored in the index
				indexedUID, err := keys.Decode(indexValue)
				if err != nil {
					continue // Skip if we can't decode
				}

				// Compare the indexed UID with the UID we're trying to delete
				if bytes.Equal(indexedUID.Encode(), id.Encode()) {
					// This index points to our UID, delete it
					if err := txn.Delete(indexKey); err != nil {
						return err
					}
					break
				}
			}

			// Find and delete the index entry that points to this UID
			indexKeys2, err := txn.Keys(r.idxprefix)
			if err != nil {
				return err
			}

			// Iterate through all index keys to find which one points to this UID
			for _, indexKey := range indexKeys2 {
				// Read the value stored in this index key (it contains the UID)
				indexValue, err := txn.Get(indexKey)
				if err != nil {
					continue // Skip if we can't read this index key
				}

				// Decode the UID stored in the index
				indexedUID, err := keys.Decode(indexValue)
				if err != nil {
					continue // Skip if we can't decode
				}

				// Compare the indexed UID with the UID we're trying to delete
				if bytes.Equal(indexedUID.Encode(), id.Encode()) {
					// This index points to our UID, delete it
					if err := txn.Delete(indexKey); err != nil {
						return err
					}
					break
				}
			}

			// Delete the main resource key
			key := id.EncodePrefixed(r.prefix)
			if err := txn.Delete(key); err != nil {
				return err
			}
		case keys.TagName:
			idxKey := id.EncodePrefixed(r.iprefix)
			idxItem, err := txn.Get(idxKey)
			if err != nil {
				return err
			}
			uid, err := keys.Decode(idxItem)
			if err != nil {
				return err
			}
			// Delete the main resource
			if err := txn.Delete(uid.EncodePrefixed(r.prefix)); err != nil {
				return err
			}
			// Delete the index key
			if err := txn.Delete(idxKey); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported id tag: %v", id.Tag())
		}

		return nil
	})
	return err
}

func (r Repo[T]) Update(ctx context.Context, id keys.ID, resource T) (T, error) {
	var res T
	err := r.db.Update(ctx, func(txn Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Fetch existing resource
		existing, err := r.Get(ctx, id)
		if err != nil {
			return err
		}

		// Preserve read-only fields from existing resource
		existingMeta := existing.GetMeta()
		resource.GetMeta().Uid = existingMeta.Uid
		resource.GetMeta().Created = existingMeta.Created
		resource.GetMeta().Updated = timestamppb.Now()
		resource.GetMeta().ResourceVersion = existingMeta.ResourceVersion + 1
		resource.GetMeta().Generation = existingMeta.Generation + 1

		// Marshal and save
		b, err := proto.Marshal(resource)
		if err != nil {
			return err
		}

		switch id.Tag() {
		case keys.TagUID:
			err := r.updateByUID(ctx, id, b)
			if err != nil {
				return err
			}

			res, err = r.getByUID(ctx, id)
			if err != nil {
				return err
			}
		case keys.TagName:
			err := r.updateByName(ctx, id, b)
			if err != nil {
				return err
			}

			res, err = r.getByName(ctx, id)
			if err != nil {
				return err
			}
		case keys.TagIdx:
			err := r.updateByIdx(ctx, id, b)
			if err != nil {
				return err
			}
			res, err = r.getByIndex(ctx, id)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported id tag: %v", id.Tag())
		}
		return nil
	})
	return res, err
}

func (r Repo[T]) updateByName(ctx context.Context, id keys.ID, b []byte) error {
	return r.db.Update(ctx, func(txn Txn) error {
		idxItem, err := txn.Get(id.EncodePrefixed(r.iprefix))
		if err != nil {
			return err
		}

		uid, err := keys.Decode(idxItem)
		if err != nil {
			return err
		}

		if err := txn.Set(uid.EncodePrefixed(r.prefix), b); err != nil {
			return err
		}

		return nil
	})
}

func (r Repo[T]) updateByIdx(ctx context.Context, id keys.ID, b []byte) error {
	return r.db.Update(ctx, func(txn Txn) error {
		idxItem, err := txn.Get(id.EncodePrefixed(r.idxprefix))
		if err != nil {
			return err
		}

		uid, err := keys.Decode(idxItem)
		if err != nil {
			return err
		}

		if err := txn.Set(uid.EncodePrefixed(r.prefix), b); err != nil {
			return err
		}

		return nil
	})
}

func (r Repo[T]) updateByUID(ctx context.Context, id keys.ID, b []byte) error {
	return r.db.Update(ctx, func(txn Txn) error {
		_, err := txn.Get(id.EncodePrefixed(r.prefix))
		if err != nil {
			return err
		}
		if err := txn.Set(id.EncodePrefixed(r.prefix), b); err != nil {
			return err
		}
		return nil
	})
}
