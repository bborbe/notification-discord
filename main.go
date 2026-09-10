// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bborbe/collection"
	"github.com/bborbe/cqrs/base"
	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	libkafka "github.com/bborbe/kafka"
	libkv "github.com/bborbe/kv"
	"github.com/bborbe/notification-discord/pkg"
	"github.com/bborbe/notification-discord/pkg/factory"
	"github.com/bborbe/notification/db"
	libfactory "github.com/bborbe/notification/factory"
	libmetrics "github.com/bborbe/notification/metrics"
	"github.com/bborbe/run"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
	libtime "github.com/bborbe/time"
	"github.com/bwmarrin/discordgo"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serviceName = "core-discord-controller"

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN       string             `required:"true"  arg:"sentry-dsn"        env:"SENTRY_DSN"        usage:"SentryDSN"                             display:"length"`
	SentryProxy     string             `required:"false" arg:"sentry-proxy"      env:"SENTRY_PROXY"      usage:"Sentry Proxy"`
	Listen          string             `required:"true"  arg:"listen"            env:"LISTEN"            usage:"address to listen to"`
	KafkaBrokers    libkafka.Brokers   `required:"true"  arg:"kafka-brokers"     env:"KAFKA_BROKERS"     usage:"Comma separated list of Kafka brokers"`
	KafkaGroup      libkafka.Group     `required:"true"  arg:"kafka-group"       env:"KAFKA_GROUP"       usage:"Kafka consumer group"`
	BatchSize       libkafka.BatchSize `required:"true"  arg:"batch-size"        env:"BATCH_SIZE"        usage:"batch consume size"                                     default:"1"`
	DataDir         string             `required:"true"  arg:"datadir"           env:"DATADIR"           usage:"data directory"`
	NoSync          bool               `required:"true"  arg:"no-sync"           env:"NO_SYNC"           usage:"no sync"                                                default:"false"`
	DiscordToken    string             `required:"true"  arg:"discord-token"     env:"DISCORD_TOKEN"     usage:"Discord token"                         display:"length"`
	DiscordServerID string             `required:"true"  arg:"discord-server-id" env:"DISCORD_SERVER_ID" usage:"Discord server id"`
	Branch          base.Branch        `required:"true"  arg:"branch"            env:"BRANCH"            usage:"branch"`
	BuildGitCommit  string             `required:"false" arg:"build-git-commit"  env:"BUILD_GIT_COMMIT"  usage:"Build Git commit hash"                                  default:"none"`
	BuildDate       *libtime.DateTime  `required:"false" arg:"build-date"        env:"BUILD_DATE"        usage:"Build timestamp (RFC3339)"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	libmetrics.NewBuildInfoMetrics().SetBuildInfo(a.BuildDate)

	saramaClientProvider, err := libkafka.NewSaramaClientProviderByType(
		ctx,
		libkafka.SaramaClientProviderTypeReused,
		a.KafkaBrokers,
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "create sarama client provider failed")
	}
	defer saramaClientProvider.Close()

	syncProducer, err := libfactory.NewSyncProducerWithName(
		ctx,
		a.KafkaBrokers,
		serviceName,
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "create sync producer failed")
	}
	defer syncProducer.Close()

	db, err := db.OpenBoltDB(ctx, a.DataDir, a.NoSync)
	if err != nil {
		return errors.Wrapf(ctx, err, "open db failed")
	}
	defer db.Close()

	discordSession, err := discordgo.New(fmt.Sprintf("Bot %s", a.DiscordToken))
	if err != nil {
		return errors.Wrapf(ctx, err, "new bot failed")
	}
	if err := discordSession.Open(); err != nil {
		return errors.Wrapf(ctx, err, "open session failed")
	}
	defer discordSession.Close()

	trigger := run.NewTrigger()

	return service.Run(
		ctx,
		a.importChannels(db, discordSession, trigger),
		run.Triggered(
			a.createCommandConsumer(
				saramaClientProvider,
				syncProducer,
				db,
				discordSession,
				sentryClient,
			),
			trigger.Done(),
		),
		a.createHTTPServer(db, discordSession),
	)
}

func (a *application) createCommandConsumer(
	saramaClientProvider libkafka.SaramaClientProvider,
	syncProducer libkafka.SyncProducer,
	db libkv.DB,
	discordSession *discordgo.Session,
	sentryClient libsentry.Client,
) run.Func {
	return factory.CreateCommandConsumer(
		db,
		saramaClientProvider,
		syncProducer,
		libkafka.Group(a.KafkaGroup),
		a.Branch,
		a.BatchSize,
		discordSession,
		pkg.ServerID(a.DiscordServerID),
		sentryClient,
	)
}

func (a *application) importChannels(
	db libkv.DB,
	session *discordgo.Session,
	trigger run.Fire,
) run.Func {
	return func(ctx context.Context) error {
		streamer := pkg.NewChannelStreamer(session, pkg.ServerID(a.DiscordServerID))
		store := pkg.NewChannelStore(db)
		var counter uint64
		err := collection.ChannelFnMap(
			ctx,
			streamer.Stream,
			func(ctx context.Context, channel pkg.Channel) error {
				counter++
				return store.Add(ctx, channel.Name, channel.ID)
			},
		)
		if err != nil {
			return errors.Wrapf(ctx, err, "map failed")
		}
		glog.V(2).Infof("imported %d channels => fire trigger", counter)
		trigger.Fire()
		<-ctx.Done()
		return ctx.Err()
	}
}

func (a *application) createHTTPServer(
	db libkv.DB,
	discordSession *discordgo.Session,
) run.Func {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		router := mux.NewRouter()
		router.Path("/healthz").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/readiness").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/metrics").Handler(promhttp.Handler())
		router.Path("/resetdb").Handler(libkv.NewResetHandler(db, cancel))
		router.Path("/resetbucket/{BucketName}").Handler(libkv.NewResetBucketHandler(db, cancel))
		router.Path("/setloglevel/{level}").Handler(libfactory.CreateSetLoglevelHandler(ctx))
		router.Path("/offsetmanager").Handler(libfactory.CreateOffsetManagerHandler(db, cancel))
		router.Path("/sendmessage").
			Handler(factory.CreateSendTestMessageHandler(db, discordSession, pkg.ServerID(a.DiscordServerID)))
		router.Path("/api/channel/list").
			Handler(factory.CreateListSessionChannelsHandler(discordSession, pkg.ServerID(a.DiscordServerID)))
		router.Path("/api/channel/exists/{id:.+}").
			Handler(factory.CreateExistsHandler(discordSession))
		router.Path("/store/channel/list").Handler(factory.CreateListStoreChannelsHandler(db))
		router.Path("/store/channel/list").Handler(factory.CreateListStoreChannelsHandler(db))
		router.Path("/channel/update").
			Handler(factory.CreateUpdateChannelHandler(db, discordSession, pkg.ServerID(a.DiscordServerID)))

		glog.V(2).Infof("starting http server listen on %s", a.Listen)
		return libhttp.NewServer(
			a.Listen,
			router,
		).Run(ctx)
	}
}
