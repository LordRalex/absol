package hyperlinkscanner

import "testing"

func Test_refreshList(t *testing.T) {
	t.Run("Test refreshing bad links", func(t *testing.T) {
		err := refreshList()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		if len(badLinks) == 0 {
			t.Error("No links were found from the API")
			t.Fail()
		}
	})
}
