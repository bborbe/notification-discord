// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"net/http"

	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	"github.com/golang/glog"

	"github.com/bborbe/notification-discord/pkg"
	"github.com/bborbe/notification/discord"
)

func NewSendTestMessageHandler(
	channelIDFinder pkg.ChannelIDProvider,
	messageSender pkg.MessageSender,
) libhttp.WithError {
	return libhttp.WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			channelName := discord.ChannelName(req.FormValue("channelName"))
			if err := channelName.Validate(ctx); err != nil {
				channelName = "test"
			}
			message := discord.Message(req.FormValue("message"))
			if message == "" {
				message = "hello world"
			}
			glog.V(2).Infof("try send message(%s) to channel(%s)", message, channelName)

			channelID, err := channelIDFinder.FindOrCreate(ctx, channelName)
			if err != nil {
				return errors.Wrapf(ctx, err, "find failed")
			}
			glog.V(2).Infof("find channel %s returned: %+v", channelName, channelID)

			if err := messageSender.Send(ctx, *channelID, message); err != nil {
				return errors.Wrapf(ctx, err, "send message failed")
			}
			_, _ = libhttp.WriteAndGlog(
				resp,
				"send message to channel(id:%s name:%s) completed",
				channelID,
				channelName,
			)
			return nil
		},
	)
}
