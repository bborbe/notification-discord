// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
)

type ChannelIDs []ChannelID

func (c ChannelIDs) Contains(channelID ChannelID) bool {
	for _, cc := range c {
		if cc == channelID {
			return true
		}
	}
	return false
}

type ChannelID string

func (c ChannelID) Validate(ctx context.Context) error {
	if len(c) == 0 {
		return errors.Wrapf(ctx, validation.Error, "channel id empty")
	}
	return nil
}

func (c ChannelID) String() string {
	return string(c)
}

func (c ChannelID) Ptr() *ChannelID {
	return &c
}

func (c ChannelID) Bytes() []byte {
	return []byte(c)
}
