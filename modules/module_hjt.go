//go:build modules.hjt || modules.all
// +build modules.hjt modules.all

package modules

import "github.com/lordralex/absol/modules/hjt"

func init() {
	Add(&hjt.Module{})
}
