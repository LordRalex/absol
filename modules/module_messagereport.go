//go:build modules.messagereport || modules.all
// +build modules.messagereport modules.all

package modules

import "github.com/lordralex/absol/modules/messagereport"

func init() {
	Add(&messagereport.Module{})
}
