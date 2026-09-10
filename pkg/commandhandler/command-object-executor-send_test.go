// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commandhandler_test

import (
	"context"

	"github.com/bborbe/cqrs/base"
	"github.com/bborbe/cqrs/cdb"
	cqrsiam "github.com/bborbe/cqrs/iam"
	iammocks "github.com/bborbe/cqrs/mocks"
	"github.com/bborbe/errors"
	libkv "github.com/bborbe/kv"
	libkvmocks "github.com/bborbe/kv/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/notification-discord/pkg"
	"github.com/bborbe/notification-discord/pkg/commandhandler"
	command "github.com/bborbe/notification/command/discord"
	"github.com/bborbe/notification/discord"
)

var _ = Describe("SendCommandObjectExecutor", func() {
	var (
		ctx                 context.Context
		tx                  *libkvmocks.Tx
		permissionChecker   *iammocks.IAMPermissionChecker
		messageSender       pkg.MessageSenderFunc
		channelIDProviderTx pkg.ChannelIDProviderTxFunc

		executor      cdb.CommandObjectExecutorTx
		commandObject cdb.CommandObject

		sendCalled    bool
		sendErr       error
		findCreateErr error
	)

	BeforeEach(func() {
		ctx = context.Background()
		tx = &libkvmocks.Tx{}
		permissionChecker = &iammocks.IAMPermissionChecker{}
		sendCalled = false
		sendErr = nil
		findCreateErr = nil

		// Default: grant permission
		permissionChecker.CheckReturns(nil)

		channelID := pkg.ChannelID("test-channel-id")
		messageSender = func(ctx context.Context, chID pkg.ChannelID, message discord.Message) error {
			sendCalled = true
			return sendErr
		}
		channelIDProviderTx = func(ctx context.Context, tx libkv.Tx, name discord.ChannelName) (*pkg.ChannelID, error) {
			if findCreateErr != nil {
				return nil, findCreateErr
			}
			return &channelID, nil
		}

		commandObject = cdb.CommandObject{
			Command: base.Command{
				ID:        "test-command-id",
				Operation: command.SendCommandOperation,
				Initiator: "test-user",
				Data: base.Event{
					"channelName": "test-channel",
					"message":     "Test Message",
				},
			},
		}
	})

	JustBeforeEach(func() {
		executor = commandhandler.NewSendCommandObjectExecutor(
			permissionChecker,
			messageSender,
			channelIDProviderTx,
		)
	})

	Context("when sending message succeeds", func() {
		It("should execute successfully", func() {
			eventID, event, err := executor.HandleCommand(ctx, tx, commandObject)

			Expect(err).NotTo(HaveOccurred())
			Expect(eventID).NotTo(BeNil())
			Expect(*eventID).To(Equal(base.EventID("test-command-id")))
			Expect(event).To(Equal(commandObject.Command.Data))
			Expect(sendCalled).To(BeTrue())
		})
	})

	Context("when message sender fails", func() {
		BeforeEach(func() {
			sendErr = errors.New(ctx, "send failed")
		})

		It("should return error", func() {
			eventID, event, err := executor.HandleCommand(ctx, tx, commandObject)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("send update failed"))
			Expect(eventID).To(BeNil())
			Expect(event).To(BeNil())
		})
	})

	Context("when channel provider fails", func() {
		BeforeEach(func() {
			findCreateErr = errors.New(ctx, "channel not found")
		})

		It("should return error", func() {
			eventID, event, err := executor.HandleCommand(ctx, tx, commandObject)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("get channelid failed"))
			Expect(eventID).To(BeNil())
			Expect(event).To(BeNil())
			Expect(sendCalled).To(BeFalse())
		})
	})

	Context("when command data is invalid", func() {
		BeforeEach(func() {
			commandObject.Command.Data = base.Event{
				"channelName": "",
			}
		})

		It("should return validation error", func() {
			eventID, event, err := executor.HandleCommand(ctx, tx, commandObject)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("validate sendCommand failed"))
			Expect(eventID).To(BeNil())
			Expect(event).To(BeNil())
			Expect(sendCalled).To(BeFalse())
		})
	})

	Context("when permission is denied", func() {
		BeforeEach(func() {
			permissionChecker.CheckReturns(cqrsiam.PermissionDeniedError)
		})

		It("should return permission denied error", func() {
			eventID, event, err := executor.HandleCommand(ctx, tx, commandObject)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("permission denied"))
			Expect(eventID).To(BeNil())
			Expect(event).To(BeNil())

			Expect(permissionChecker.CheckCallCount()).To(Equal(1))
			Expect(sendCalled).To(BeFalse())
		})
	})

	Context("operation flags", func() {
		It("should have correct operation", func() {
			Expect(executor.CommandOperation()).To(Equal(command.SendCommandOperation))
		})

		It("should have send result enabled", func() {
			Expect(executor.SendResultEnabled()).To(BeTrue())
		})
	})
})
