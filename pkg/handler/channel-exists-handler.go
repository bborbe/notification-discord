// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"net/http"

	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	"github.com/bborbe/notification-discord/pkg"
	"github.com/gorilla/mux"
)

func NewChannelExistHandler(checker pkg.ChannelExistsChecker) http.Handler {
	return libhttp.NewErrorHandler(
		libhttp.WithErrorFunc(
			func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
				vars := mux.Vars(req)
				channelID := pkg.ChannelID(vars["id"])
				exists, err := checker.Exists(ctx, channelID)
				if err != nil {
					return errors.Wrapf(ctx, err, "exists failed")
				}
				_, _ = libhttp.WriteAndGlog(
					resp,
					"channel with id %s exists = %v",
					channelID,
					exists,
				)
				return nil
			},
		),
	)
}
