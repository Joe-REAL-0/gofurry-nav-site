package controller

import (
	"errors"
	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/jackc/pgx/v5/pgconn"
	"reflect"
	"strings"
	"testing"
)

func TestCollectionValidation(t *testing.T) {
	for _, code := range []string{"a", "furry-2026", strings.Repeat("a", 64)} {
		if !collectionCodePattern.MatchString(code) {
			t.Fatal(code)
		}
	}
	for _, code := range []string{"", "Upper", "-a", "a-", "a--b", "分区", "a_b"} {
		if collectionCodePattern.MatchString(code) {
			t.Fatal(code)
		}
	}
	content := models.CollectionContent{Name: " 测试 ", NameEn: " Test "}
	if e := normalizeCollectionContent(&content); e != nil || content.Name != "测试" {
		t.Fatal(content, e)
	}
	for _, in := range []models.CollectionContent{{Name: "", NameEn: "Test"}, {Name: "测试", NameEn: " "}, {Name: strings.Repeat("测", 161), NameEn: "Test"}, {Name: "测", NameEn: "Test", Info: strings.Repeat("测", 501)}} {
		if normalizeCollectionContent(&in) == nil {
			t.Fatal(in)
		}
	}
	content.Info = "简介"
	content.InfoEn = "Description"
	if requirePublishable(content, 2) != nil || requirePublishable(content, 1) == nil {
		t.Fatal("readiness")
	}
	ids, e := canonicalCollectionIDs([]int64{4, 2, 4, 1})
	if e != nil || !reflect.DeepEqual(ids, []int64{1, 2, 4}) {
		t.Fatal(ids, e)
	}
	if _, e = canonicalCollectionIDs([]int64{0}); e == nil {
		t.Fatal("zero ID accepted")
	}
	if _, e = canonicalCollectionIDs([]int64{-1}); e == nil {
		t.Fatal("negative ID accepted")
	}
}
func TestCollectionTransitionMatrix(t *testing.T) {
	valid := map[string]string{"draft/publish": "published", "published/unpublish": "draft", "draft/archive": "archived", "published/archive": "archived", "archived/restore": "draft"}
	for _, status := range []string{"draft", "published", "archived"} {
		for _, action := range []string{"publish", "unpublish", "archive", "restore"} {
			next, e := collectionTransition(status, action)
			want, ok := valid[status+"/"+action]
			if ok && (e != nil || next != want) || !ok && e == nil {
				t.Fatalf("%s/%s: %s %v", status, action, next, e)
			}
		}
	}
}
func TestCollectionHomeValidationAndRevision(t *testing.T) {
	slots := make([]models.CollectionPlacement, 5)
	for i := range slots {
		slots[i].Slot = int16(i + 1)
	}
	valid := models.ReplaceCollectionHome{Revision: "revision", Slots: slots}
	if e := validateCollectionHome(valid); e != nil {
		t.Fatal(e)
	}
	invalid := valid
	invalid.Slots = slots[:4]
	if validateCollectionHome(invalid) == nil {
		t.Fatal("incomplete")
	}
	invalid = valid
	invalid.Revision = ""
	if validateCollectionHome(invalid) == nil {
		t.Fatal("missing revision")
	}
	id := int64(12)
	slots[0].CollectionID = &id
	slots[1].CollectionID = &id
	if validateCollectionHome(valid) == nil {
		t.Fatal("duplicate collection")
	}
	slots[1].CollectionID = nil
	slots[4].Slot = 1
	if validateCollectionHome(valid) == nil {
		t.Fatal("duplicate slot")
	}
	slots[4].Slot = 6
	if validateCollectionHome(valid) == nil {
		t.Fatal("sixth slot")
	}
	a := collectionRevision([5]int64{12, 0, 0, 0, 0})
	if a != collectionRevision([5]int64{12, 0, 0, 0, 0}) || a == collectionRevision([5]int64{0, 12, 0, 0, 0}) {
		t.Fatal("placement revision")
	}
}
func TestCollectionErrorRedaction(t *testing.T) {
	if e := collectionError(&pgconn.PgError{Code: "23505", Message: "secret SQL"}); e.GetHTTPStatus() != 409 || strings.Contains(e.Error(), "secret") {
		t.Fatal(e)
	}
	if e := collectionError(errors.New("database secret")); e.GetHTTPStatus() != 500 || strings.Contains(e.Error(), "secret") {
		t.Fatal(e)
	}
}
