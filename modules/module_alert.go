//go:build modules.alert || modules.all

package modules

import "github.com/lordralex/absol/modules/alert"

func init() {
	Add(&alert.Module{})
}
