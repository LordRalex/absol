//go:build hjt || all
// +build hjt all

package modules

import "github.com/lordralex/absol/modules/hjt"

func init() {
	Add(&hjt.Module{})
}
