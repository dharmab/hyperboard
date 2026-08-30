package api

import (
	"context"
	"errors"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/storage"
)

type faultInjectingSQLStore struct {
	store.SQLStore
	pingErr              error
	convertTagToAliasErr error
}

func (s faultInjectingSQLStore) Ping(ctx context.Context) error {
	if s.pingErr != nil {
		return s.pingErr
	}
	return s.SQLStore.Ping(ctx)
}

func (s faultInjectingSQLStore) ConvertTagToAlias(ctx context.Context, sourceName, targetName string) (*store.ConvertTagToAliasResult, error) {
	if s.convertTagToAliasErr != nil {
		return nil, s.convertTagToAliasErr
	}
	return s.SQLStore.ConvertTagToAlias(ctx, sourceName, targetName)
}

type faultInjectingMediaStore struct {
	storage.MediaStore
	pingErr      error
	failUploadAt int
	uploads      int
	failDeleteAt int
	deletes      int
	deletedKeys  []string
}

func (s *faultInjectingMediaStore) Ping(ctx context.Context) error {
	if s.pingErr != nil {
		return s.pingErr
	}
	return s.MediaStore.Ping(ctx)
}

func (s *faultInjectingMediaStore) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	s.uploads++
	if s.uploads == s.failUploadAt {
		return "", errors.New("injected upload failure")
	}
	return s.MediaStore.Upload(ctx, key, data, contentType)
}

func (s *faultInjectingMediaStore) Delete(ctx context.Context, key string) error {
	s.deletes++
	s.deletedKeys = append(s.deletedKeys, key)
	if s.deletes == s.failDeleteAt {
		return errors.New("injected delete failure")
	}
	return s.MediaStore.Delete(ctx, key)
}
