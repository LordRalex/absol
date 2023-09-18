//go:build modules.polls || modules.all

package modules

import "github.com/lordralex/absol/modules/polls"

func init() {
	Add(&polls.Module{})
}
