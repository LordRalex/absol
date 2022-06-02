//go:build twitch || all
// +build twitch all

package modules

import (
	"github.com/lordralex/absol/modules/twitch"
)

func init() {
	Add(&twitch.Module{})
}
