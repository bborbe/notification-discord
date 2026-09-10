// Copyright (c) 2025 Benjamin Borbe All rights reserved.
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

// ChannelIDProviderTx is a transaction-aware version of ChannelIDProvider
// that accepts an existing transaction instead of opening a new one.
type ChannelIDProviderTx interface {
	FindOrCreate(ctx context.Context, tx libkv.Tx, name discord.ChannelName) (*ChannelID, error)
}

// ChannelIDProviderTxFunc is a function adapter for ChannelIDProviderTx
type ChannelIDProviderTxFunc func(ctx context.Context, tx libkv.Tx, name discord.ChannelName) (*ChannelID, error)

func (c ChannelIDProviderTxFunc) FindOrCreate(
	ctx context.Context,
	tx libkv.Tx,
	name discord.ChannelName,
) (*ChannelID, error) {
	return c(ctx, tx, name)
}

// NewChannelIDProviderStoreTx creates a transaction-aware channel ID provider
// that uses the existing transaction instead of opening a new one.
func NewChannelIDProviderStoreTx(
	channelStoreTx ChannelStoreTx,
	channelIDProvider ChannelIDProvider,
	channelExistsChecker ChannelExistsChecker,
) ChannelIDProviderTx {
	return ChannelIDProviderTxFunc(
		func(ctx context.Context, tx libkv.Tx, channelName discord.ChannelName) (*ChannelID, error) {
			channelID, err := channelStoreTx.Get(ctx, tx, channelName)
			if err == nil {
				glog.V(4).Infof("found channelName(%s) in store", channelName)
				exists, err := channelExistsChecker.Exists(ctx, *channelID)
				if err != nil {
					return nil, errors.Wrapf(ctx, err, "exists failed")
				}
				if exists {
					glog.V(4).Infof("channelName(%s) exists => return", channelName)
					return channelID, nil
				}
				glog.V(2).Infof("channelName(%s) does not exist => create", channelName)
			}
			channelID, err = channelIDProvider.FindOrCreate(ctx, channelName)
			if err != nil {
				return nil, errors.Wrapf(ctx, err, "find or create channel(%s) failed", channelName)
			}
			glog.V(4).Infof("found channelID(%s) for channelName(%s)", channelID, channelName)
			if err := channelStoreTx.Add(ctx, tx, channelName, *channelID); err != nil {
				return nil, errors.Wrapf(ctx, err, "add channel failed")
			}
			glog.V(4).Infof("channelName(%s) added to store", channelName)
			return channelID, nil
		},
	)
}
