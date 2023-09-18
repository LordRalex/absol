//go:build modules.cleaner || modules.all

package modules

import "github.com/lordralex/absol/modules/cleaner"

func init() {
	Add(&cleaner.Module{})
}
