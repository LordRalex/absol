//go:build cleaner || all
// +build cleaner all

package modules

import "github.com/lordralex/absol/modules/cleaner"

func init() {
	Add(&cleaner.Module{})
}
