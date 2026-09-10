// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commandhandler

import (
	"context"

	"github.com/bborbe/cqrs/base"
	"github.com/bborbe/cqrs/cdb"
	cqrsiam "github.com/bborbe/cqrs/iam"
	"github.com/bborbe/errors"
	libkv "github.com/bborbe/kv"
	"github.com/bborbe/notification-discord/pkg"
	command "github.com/bborbe/notification/command/discord"
	"github.com/bborbe/notification/iam"
	"github.com/golang/glog"
)

func NewSendCommandObjectExecutor(
	permissionChecker cqrsiam.PermissionChecker,
	messageSender pkg.MessageSender,
	channelIDProviderTx pkg.ChannelIDProviderTx,
) cdb.CommandObjectExecutorTx {
	return cdb.CommandObjectExecutorTxFunc(
		command.SendCommandOperation,
		true,
		func(ctx context.Context, tx libkv.Tx, commandObject cdb.CommandObject) (*base.EventID, base.Event, error) {
			glog.V(2).Infof("send discord message started")

			// Check permissions
			permissionCheck := iam.NewAnyPermissionCheck(
				iam.CoreDiscordSendPermission,
				iam.CoreDiscordAdminPermission,
			)
			if err := permissionChecker.Check(ctx, tx, commandObject.Command.Initiator, permissionCheck); err != nil {
				return nil, nil, errors.Wrapf(ctx, err, "permission denied")
			}

			event := commandObject.Command.Data
			var sendCommand command.SendCommand
			if err := event.MarshalInto(ctx, &sendCommand); err != nil {
				return nil, nil, errors.Wrapf(ctx, err, "marshal into sendCommand failed")
			}
			if err := sendCommand.Validate(ctx); err != nil {
				return nil, nil, errors.Wrapf(ctx, err, "validate sendCommand failed")
			}
			channelID, err := channelIDProviderTx.FindOrCreate(ctx, tx, sendCommand.ChannelName)
			if err != nil {
				return nil, nil, errors.Wrapf(ctx, err, "get channelid failed")
			}
			if err := messageSender.Send(ctx, *channelID, sendCommand.Message); err != nil {
				return nil, nil, errors.Wrapf(ctx, err, "send update failed")
			}
			glog.V(2).Infof("send message to channel(%s) completed", sendCommand.ChannelName)
			return commandObject.Command.ID.Ptr(), event, nil
		},
	)
}
