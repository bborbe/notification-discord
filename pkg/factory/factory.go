// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"net/http"
	"time"

	"github.com/bborbe/cqrs/base"
	"github.com/bborbe/cqrs/cdb"
	cqrsiam "github.com/bborbe/cqrs/iam"
	libhttp "github.com/bborbe/http"
	libkafka "github.com/bborbe/kafka"
	libkv "github.com/bborbe/kv"
	core "github.com/bborbe/notification"
	"github.com/bborbe/notification-discord/pkg"
	"github.com/bborbe/notification-discord/pkg/commandhandler"
	"github.com/bborbe/notification-discord/pkg/handler"
	"github.com/bborbe/run"
	libsentry "github.com/bborbe/sentry"
	"github.com/bwmarrin/discordgo"
)

func CreateCommandConsumer(
	db libkv.DB,
	saramaClientProvider libkafka.SaramaClientProvider,
	syncProducer libkafka.SyncProducer,
	kafkaGroup libkafka.Group,
	branch base.Branch,
	batchSize libkafka.BatchSize,
	discordSession *discordgo.Session,
	serverID pkg.ServerID,
	sentryClient libsentry.Client,
) run.Func {
	permissionChecker := cqrsiam.NewPermissionChecker(
		sentryClient,
		cqrsiam.NewPermissionCheckerMetrics(),
	)
	return cdb.RunCommandConsumerTx(
		saramaClientProvider,
		syncProducer,
		db,
		core.DiscordV1SchemaID,
		batchSize,
		base.TopicPrefixFromBranch(branch),
		false,
		24*time.Hour,
		run.NewTrigger(),
		cdb.CommandObjectExecutorTxs{
			CreateSendCommandObjectExecutor(
				permissionChecker,
				db,
				discordSession,
				serverID,
			),
		},
	)
}

func CreateSendCommandObjectExecutor(
	permissionChecker cqrsiam.PermissionChecker,
	db libkv.DB,
	discordSession *discordgo.Session,
	serverID pkg.ServerID,
) cdb.CommandObjectExecutorTx {
	return commandhandler.NewSendCommandObjectExecutor(
		permissionChecker,
		pkg.NewMessageSender(discordSession),
		CreateChannelProviderTx(
			db,
			discordSession,
			serverID,
		),
	)
}

func CreateSendTestMessageHandler(
	db libkv.DB,
	discordSession *discordgo.Session,
	serverID pkg.ServerID,
) http.Handler {
	return libhttp.NewErrorHandler(
		handler.NewSendTestMessageHandler(
			CreateChannelProvider(
				db,
				discordSession,
				serverID,
			),
			pkg.NewMessageSender(discordSession),
		),
	)
}

func CreateListSessionChannelsHandler(
	discordSession *discordgo.Session,
	serverID pkg.ServerID,
) http.Handler {
	return libhttp.NewErrorHandler(
		handler.NewChannelListHandler(
			pkg.NewChannelStreamer(
				discordSession,
				serverID,
			),
		),
	)
}

func CreateExistsHandler(discordSession *discordgo.Session) http.Handler {
	return handler.NewChannelExistHandler(
		pkg.NewChannelExistsChecker(discordSession),
	)
}

func CreateUpdateChannelHandler(
	db libkv.DB,
	discordSession *discordgo.Session,
	serverID pkg.ServerID,
) http.Handler {
	return libhttp.NewErrorHandler(
		handler.NewChannelUpdateHandler(
			pkg.NewChannelStreamer(
				discordSession,
				serverID,
			),
			pkg.NewChannelStore(db),
		),
	)
}

func CreateListStoreChannelsHandler(
	db libkv.DB,
) http.Handler {
	return libhttp.NewErrorHandler(
		handler.NewChannelListHandler(
			pkg.NewChannelStore(db),
		),
	)
}

// CreateChannelProvider creates a non-transaction-aware channel provider.
// Used for HTTP handlers that manage their own transactions.
func CreateChannelProvider(
	db libkv.DB,
	discordSession *discordgo.Session,
	serverID pkg.ServerID,
) pkg.ChannelIDProvider {
	return pkg.NewChannelIDProviderStore(
		pkg.NewChannelStore(db),
		pkg.NewChannelIDProvider(
			pkg.NewChannelIDFinder(
				pkg.NewChannelStreamer(
					discordSession,
					serverID,
				),
			),
			pkg.NewChannelCreator(
				discordSession,
				serverID,
			),
		),
		pkg.NewChannelExistsChecker(discordSession),
	)
}

// CreateChannelProviderTx creates a transaction-aware channel provider.
// Used for command handlers that already have an active transaction.
func CreateChannelProviderTx(
	db libkv.DB,
	discordSession *discordgo.Session,
	serverID pkg.ServerID,
) pkg.ChannelIDProviderTx {
	return pkg.NewChannelIDProviderStoreTx(
		pkg.NewChannelStoreTx(),
		pkg.NewChannelIDProvider(
			pkg.NewChannelIDFinder(
				pkg.NewChannelStreamer(
					discordSession,
					serverID,
				),
			),
			pkg.NewChannelCreator(
				discordSession,
				serverID,
			),
		),
		pkg.NewChannelExistsChecker(discordSession),
	)
}
