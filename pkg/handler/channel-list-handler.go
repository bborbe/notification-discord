// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bborbe/collection"
	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	"github.com/bborbe/notification-discord/pkg"
)

func NewChannelListHandler(channelStreamer pkg.ChannelStreamer) libhttp.WithError {
	return libhttp.WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			channels, err := collection.ChannelFnList(ctx, channelStreamer.Stream)
			if err != nil {
				return errors.Wrapf(ctx, err, "list channels failed")
			}
			resp.Header().Add(libhttp.ContentTypeHeaderName, libhttp.ApplicationJsonContentType)
			return json.NewEncoder(resp).Encode(channels)
		},
	)
}
