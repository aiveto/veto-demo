package desks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrdersListHonorsStatusAndLimit(t *testing.T) {
	h := New().Handler("orders")

	t.Run("all", func(t *testing.T) {
		ids := listIDs(t, h, "/orders")
		if got, want := ids, []string{"10502", "10490", "10482"}; !same(got, want) {
			t.Fatalf("ids=%v want %v", got, want)
		}
	})
	t.Run("placed", func(t *testing.T) {
		ids := listIDs(t, h, "/orders?status=placed")
		if got, want := ids, []string{"10490"}; !same(got, want) {
			t.Fatalf("ids=%v want %v", got, want)
		}
	})
	t.Run("limit", func(t *testing.T) {
		ids := listIDs(t, h, "/orders?limit=1")
		if got, want := ids, []string{"10502"}; !same(got, want) {
			t.Fatalf("ids=%v want %v", got, want)
		}
	})
}

func listIDs(t *testing.T, h http.Handler, path string) []string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+Token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.Bytes())
	}
	var page struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, len(page.Data))
	for i, row := range page.Data {
		ids[i] = row.ID
	}
	return ids
}

func same(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
