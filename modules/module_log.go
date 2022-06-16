//go:build modules.log || modules.all
// +build modules.log modules.all

package modules

import "github.com/lordralex/absol/modules/log"

func init() {
	Add(&log.Module{})
}
