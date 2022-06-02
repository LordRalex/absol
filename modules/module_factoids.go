//go:build factoids || all
// +build factoids all

package modules

import (
	"github.com/lordralex/absol/modules/factoids"
)

func init() {
	Add(&factoids.Module{})
}
