//go:build modules.pastes || modules.all
// +build modules.pastes modules.all

package modules

import (
	"github.com/lordralex/absol/modules/pastes"
)

func init() {
	Add(&pastes.Module{})
}
