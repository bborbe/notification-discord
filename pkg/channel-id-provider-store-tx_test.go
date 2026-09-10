// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"context"

	libkv "github.com/bborbe/kv"
	"github.com/bborbe/notification-discord/pkg"
	"github.com/bborbe/notification-discord/pkg/mocks"
	"github.com/bborbe/notification/db"
	"github.com/bborbe/notification/discord"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ChannelIDProviderStoreTx", func() {
	var (
		ctx                  context.Context
		database             libkv.DB
		channelName          discord.ChannelName
		channelExistsChecker *mocks.ChannelExistsChecker
		channelCreator       *mocks.ChannelCreator
		channelFinder        *mocks.ChannelIDFinder
	)

	BeforeEach(func() {
		var err error
		ctx = context.Background()
		channelName = discord.ChannelName("test-channel")

		// Use real in-memory database
		database, err = db.OpenMemoryDB(ctx)
		Expect(err).NotTo(HaveOccurred())

		channelExistsChecker = &mocks.ChannelExistsChecker{}
		channelExistsChecker.ExistsReturns(true, nil)

		channelCreator = &mocks.ChannelCreator{}
		channelCreator.CreateChannelReturns(&pkg.Channel{
			ID:   pkg.ChannelID("created-channel-id"),
			Name: channelName,
		}, nil)

		channelFinder = &mocks.ChannelIDFinder{}
		channelFinder.FindByNameReturns(pkg.Channels{}, nil)
	})

	AfterEach(func() {
		if database != nil {
			database.Close()
		}
	})

	Context("when called inside an existing transaction", func() {
		It("should work without nested transaction error", func() {
			channelStoreTx := pkg.NewChannelStoreTx()
			channelIDProvider := pkg.NewChannelIDProvider(channelFinder, channelCreator)
			provider := pkg.NewChannelIDProviderStoreTx(
				channelStoreTx,
				channelIDProvider,
				channelExistsChecker,
			)

			// Simulate being inside a transaction (like command executor does)
			var resultID *pkg.ChannelID
			err := database.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
				var err error
				resultID, err = provider.FindOrCreate(ctx, tx, channelName)
				return err
			})

			// Should succeed without nested transaction error
			Expect(err).NotTo(HaveOccurred())
			Expect(resultID).NotTo(BeNil())
			Expect(*resultID).To(Equal(pkg.ChannelID("created-channel-id")))
		})

		It("should save channel to store for future lookups", func() {
			channelStoreTx := pkg.NewChannelStoreTx()
			channelIDProvider := pkg.NewChannelIDProvider(channelFinder, channelCreator)
			provider := pkg.NewChannelIDProviderStoreTx(
				channelStoreTx,
				channelIDProvider,
				channelExistsChecker,
			)

			// First call creates channel
			err := database.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
				_, err := provider.FindOrCreate(ctx, tx, channelName)
				return err
			})
			Expect(err).NotTo(HaveOccurred())

			// Second call should find it in store
			var resultID *pkg.ChannelID
			err = database.View(ctx, func(ctx context.Context, tx libkv.Tx) error {
				var err error
				resultID, err = channelStoreTx.Get(ctx, tx, channelName)
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(*resultID).To(Equal(pkg.ChannelID("created-channel-id")))
		})
	})
})
