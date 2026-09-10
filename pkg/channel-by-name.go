// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import "strings"

type ChannelByName Channels

func (p ChannelByName) Len() int { return len(p) }

func (p ChannelByName) Less(i, j int) bool {
	return strings.Compare(p[i].Name.String(), p[j].Name.String()) < 0
}

func (p ChannelByName) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
