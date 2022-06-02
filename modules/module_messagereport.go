//go:build messagereport || all
// +build messagereport all

package modules

import "github.com/lordralex/absol/modules/messagereport"

func init() {
	Add(&messagereport.Module{})
}
