//go:build modules.twitch || modules.all
// +build modules.twitch modules.all

package modules

import (
	"github.com/lordralex/absol/modules/twitch"
)

func init() {
	Add(&twitch.Module{})
}
