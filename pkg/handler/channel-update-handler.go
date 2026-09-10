// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"net/http"

	"github.com/bborbe/collection"
	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	"github.com/bborbe/notification-discord/pkg"
	"github.com/golang/glog"
)

func NewChannelUpdateHandler(
	channelStreamer pkg.ChannelStreamer,
	channelStore pkg.ChannelStore,
) libhttp.WithError {
	return libhttp.WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			err := collection.ChannelFnMap(
				ctx,
				channelStreamer.Stream,
				func(ctx context.Context, channel pkg.Channel) error {
					if err := channelStore.Add(ctx, channel.Name, channel.ID); err != nil {
						return errors.Wrapf(ctx, err, "add channel failed")
					}
					glog.V(4).Infof("add channel(%s) with id(%s)", channel.Name, channel.ID)
					return nil
				},
			)
			if err != nil {
				return errors.Wrapf(ctx, err, "map channels failed")
			}
			_, _ = libhttp.WriteAndGlog(resp, "update channels completed")
			return nil
		},
	)
}
