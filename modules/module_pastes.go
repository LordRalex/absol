//go:build pastes || all
// +build pastes all

package modules

import (
	"github.com/lordralex/absol/modules/pastes"
)

func init() {
	Add(&pastes.Module{})
}
