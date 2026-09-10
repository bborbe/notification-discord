// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/channel-exists-checker.go --fake-name ChannelExistsChecker . ChannelExistsChecker
type ChannelExistsChecker interface {
	Exists(ctx context.Context, channelID ChannelID) (bool, error)
}

func NewChannelExistsChecker(discordSession *discordgo.Session) ChannelExistsChecker {
	return &channelExistsChecker{
		discordSession: discordSession,
	}
}

type channelExistsChecker struct {
	discordSession *discordgo.Session
}

func (c *channelExistsChecker) Exists(ctx context.Context, channelID ChannelID) (bool, error) {
	if _, err := c.discordSession.ChannelMessages(channelID.String(), 1, "", "", ""); err != nil {
		var restError *discordgo.RESTError
		if errors.As(err, &restError) {
			return restError.Message.Code != discordgo.ErrCodeUnknownChannel, nil
		}
		glog.V(2).Infof("error Type: %T", err)
		return false, err
	}
	return true, nil
}
