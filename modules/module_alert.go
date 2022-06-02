//go:build alert || all
// +build alert all

package modules

import "github.com/lordralex/absol/modules/alert"

func init() {
	Add(&alert.Module{})
}
