//go:build mcping || all
// +build mcping all

package modules

import "github.com/lordralex/absol/modules/mcping"

func init() {
	Add(&mcping.Module{})
}
