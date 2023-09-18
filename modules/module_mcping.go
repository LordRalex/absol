//go:build modules.mcping || modules.all

package modules

import "github.com/lordralex/absol/modules/mcping"

func init() {
	Add(&mcping.Module{})
}
