// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	libkv "github.com/bborbe/kv"

	"github.com/bborbe/notification/discord"
)

type ChannelStore interface {
	Add(ctx context.Context, name discord.ChannelName, channelID ChannelID) error
	Get(ctx context.Context, name discord.ChannelName) (*ChannelID, error)
	Stream(ctx context.Context, ch chan<- Channel) error
	Purge(ctx context.Context) error
	Delete(ctx context.Context, name discord.ChannelName) error
}

func NewChannelStore(db libkv.DB) ChannelStore {
	return &channelStore{
		db:      db,
		storeTx: NewChannelStoreTx(),
	}
}

type channelStore struct {
	db      libkv.DB
	storeTx ChannelStoreTx
}

func (c *channelStore) Purge(ctx context.Context) error {
	return c.db.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
		return c.storeTx.Purge(ctx, tx)
	})
}

func (c *channelStore) Delete(ctx context.Context, name discord.ChannelName) error {
	return c.db.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
		return c.storeTx.Delete(ctx, tx, name)
	})
}

func (c *channelStore) Stream(ctx context.Context, ch chan<- Channel) error {
	return c.db.View(ctx, func(ctx context.Context, tx libkv.Tx) error {
		return c.storeTx.Stream(ctx, tx, ch)
	})
}

func (c *channelStore) Add(
	ctx context.Context,
	name discord.ChannelName,
	channelID ChannelID,
) error {
	return c.db.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
		return c.storeTx.Add(ctx, tx, name, channelID)
	})
}

func (c *channelStore) Get(ctx context.Context, name discord.ChannelName) (*ChannelID, error) {
	var result *ChannelID
	err := c.db.View(ctx, func(ctx context.Context, tx libkv.Tx) error {
		var err error
		result, err = c.storeTx.Get(ctx, tx, name)
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "view failed")
	}
	return result, nil
}
