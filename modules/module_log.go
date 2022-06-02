//go:build log || all
// +build log all

package modules

import "github.com/lordralex/absol/modules/log"

func init() {
	Add(&log.Module{})
}
