// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	libkv "github.com/bborbe/kv"
	"github.com/golang/glog"

	"github.com/bborbe/notification/discord"
)

type ChannelStoreTx interface {
	Add(ctx context.Context, tx libkv.Tx, name discord.ChannelName, channelID ChannelID) error
	Get(ctx context.Context, tx libkv.Tx, name discord.ChannelName) (*ChannelID, error)
	Stream(ctx context.Context, tx libkv.Tx, ch chan<- Channel) error
	Purge(ctx context.Context, tx libkv.Tx) error
	Delete(ctx context.Context, tx libkv.Tx, name discord.ChannelName) error
}

func NewChannelStoreTx() ChannelStoreTx {
	return &channelStoreTx{
		bucketName: libkv.NewBucketName("channel-store"),
	}
}

type channelStoreTx struct {
	bucketName libkv.BucketName
}

func (c *channelStoreTx) Purge(ctx context.Context, tx libkv.Tx) error {
	bucket, err := tx.Bucket(ctx, c.bucketName)
	if err != nil {
		glog.V(3).Infof("bucket %s not found", c.bucketName)
		return nil
	}
	it := bucket.Iterator()
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := bucket.Delete(ctx, it.Item().Key()); err != nil {
				return errors.Wrapf(ctx, err, "delete failed")
			}
		}
	}
	return nil
}

func (c *channelStoreTx) Delete(ctx context.Context, tx libkv.Tx, name discord.ChannelName) error {
	bucket, err := tx.Bucket(ctx, c.bucketName)
	if err != nil {
		glog.V(3).Infof("bucket %s not found", c.bucketName)
		return nil
	}
	if err := bucket.Delete(ctx, name.Bytes()); err != nil {
		return errors.Wrapf(ctx, err, "delete failed")
	}
	return nil
}

func (c *channelStoreTx) Stream(ctx context.Context, tx libkv.Tx, ch chan<- Channel) error {
	bucket, err := tx.Bucket(ctx, c.bucketName)
	if err != nil {
		glog.V(3).Infof("bucket %s not found", c.bucketName)
		return nil
	}
	it := bucket.Iterator()
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			item := it.Item()
			err := item.Value(func(v []byte) error {
				channel := Channel{
					ID:   ChannelID(item.Key()),
					Name: discord.ChannelName(v),
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case ch <- channel:
					return nil
				}
			})
			if err != nil {
				return errors.Wrapf(ctx, err, "value failed")
			}
		}
	}
	return nil
}

func (c *channelStoreTx) Add(
	ctx context.Context,
	tx libkv.Tx,
	name discord.ChannelName,
	channelID ChannelID,
) error {
	bucket, err := tx.CreateBucketIfNotExists(ctx, c.bucketName)
	if err != nil {
		return errors.Wrapf(ctx, err, "create bucket failed")
	}
	if err := bucket.Put(ctx, name.Bytes(), channelID.Bytes()); err != nil {
		return errors.Wrapf(ctx, err, "add failed")
	}
	return nil
}

func (c *channelStoreTx) Get(
	ctx context.Context,
	tx libkv.Tx,
	name discord.ChannelName,
) (*ChannelID, error) {
	var result *ChannelID
	bucket, err := tx.Bucket(ctx, c.bucketName)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "get bucket failed")
	}
	item, err := bucket.Get(ctx, name.Bytes())
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "get failed")
	}
	err = item.Value(func(val []byte) error {
		if len(val) == 0 {
			return ChannelNotFoundError
		}
		result = ChannelID(val).Ptr()
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "view failed")
	}
	return result, nil
}
