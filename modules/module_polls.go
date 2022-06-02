//go:build polls || all
// +build polls all

package modules

import "github.com/lordralex/absol/modules/polls"

func init() {
	Add(&polls.Module{})
}
