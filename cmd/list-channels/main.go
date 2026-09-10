// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/bborbe/collection"
	"github.com/bborbe/errors"
	"github.com/bborbe/notification-discord/pkg"
	libmetrics "github.com/bborbe/notification/metrics"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
	libtime "github.com/bborbe/time"
	"github.com/bwmarrin/discordgo"
	"github.com/golang/glog"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN       string            `required:"true"  arg:"sentry-dsn"        env:"SENTRY_DSN"        usage:"SentryDSN"                 display:"length"`
	SentryProxy     string            `required:"false" arg:"sentry-proxy"      env:"SENTRY_PROXY"      usage:"Sentry Proxy"`
	DiscordToken    string            `required:"true"  arg:"discord-token"     env:"DISCORD_TOKEN"     usage:"Discord token"             display:"length"`
	DiscordServerID string            `required:"true"  arg:"discord-server-id" env:"DISCORD_SERVER_ID" usage:"Discord server id"`
	BuildGitCommit  string            `required:"false" arg:"build-git-commit"  env:"BUILD_GIT_COMMIT"  usage:"Build Git commit hash"                      default:"none"`
	BuildDate       *libtime.DateTime `required:"false" arg:"build-date"        env:"BUILD_DATE"        usage:"Build timestamp (RFC3339)"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	libmetrics.NewBuildInfoMetrics().SetBuildInfo(a.BuildDate)

	discordSession, err := discordgo.New(fmt.Sprintf("Bot %s", a.DiscordToken))
	if err != nil {
		return errors.Wrapf(ctx, err, "new bot failed")
	}
	if err := discordSession.Open(); err != nil {
		return errors.Wrapf(ctx, err, "open session failed")
	}
	defer discordSession.Close()

	channelStreamer := pkg.NewChannelStreamer(
		discordSession,
		pkg.ServerID(a.DiscordServerID),
	)

	list, err := collection.ChannelFnList(ctx, channelStreamer.Stream)
	if err != nil {
		return errors.Wrapf(ctx, err, "create liste failed")
	}
	sort.Sort(pkg.ChannelByName(list))

	for _, channel := range list {
		glog.V(2).Infof("%+v", channel)
	}
	return nil
}
