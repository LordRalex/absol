//go:build modules.factoids || modules.all
// +build modules.factoids modules.all

package modules

import (
	"github.com/lordralex/absol/modules/factoids"
)

func init() {
	Add(&factoids.Module{})
}
