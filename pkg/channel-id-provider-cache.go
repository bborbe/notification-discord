// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"sync"

	"github.com/bborbe/errors"

	"github.com/bborbe/notification/discord"
)

func NewChannelIDProviderCache(
	channelIDProvider ChannelIDProvider,
) ChannelIDProvider {
	var cache = make(map[discord.ChannelName]ChannelID)
	var mux sync.Mutex
	return ChannelIDProviderFunc(
		func(ctx context.Context, name discord.ChannelName) (*ChannelID, error) {
			mux.Lock()
			defer mux.Unlock()
			id, found := cache[name]
			if found {
				return &id, nil
			}
			channelID, err := channelIDProvider.FindOrCreate(ctx, name)
			if err != nil {
				return nil, errors.Wrapf(ctx, err, "find or create channel '%s' failed", name)
			}
			cache[name] = *channelID
			return channelID, nil
		},
	)
}
