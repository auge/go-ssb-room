package admin

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ssb-ngi-pointer/go-ssb-room/admindb"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/router"
	refs "go.mindeco.de/ssb-refs"
)

func TestAllowListEmpty(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)

	url, err := ts.Router.Get(router.AdminAllowListOverview).URL()
	a.Nil(err)

	html, resp := ts.Client.GetHTML(url.String(), nil)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	assertLocalized(t, html, []localizedElement{
		{"#welcome", "AdminAllowListWelcome"},
		{"title", "AdminAllowListTitle"},
		{"#allowListCount", "ListCountPlural"},
	})
}

func TestAllowListAdd(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)

	listURL, err := ts.Router.Get(router.AdminAllowListOverview).URL()
	a.NoError(err)

	html, resp := ts.Client.GetHTML(listURL.String(), nil)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	formSelection := html.Find("form#add-entry")
	a.EqualValues(1, formSelection.Length())

	method, ok := formSelection.Attr("method")
	a.True(ok, "form has method set")
	a.Equal("POST", method)

	action, ok := formSelection.Attr("action")
	a.True(ok, "form has action set")

	addURL, err := ts.Router.Get(router.AdminAllowListAdd).URL()
	a.NoError(err)

	a.Equal(addURL.String(), action)

	inputSelection := formSelection.Find("input[type=text]")
	a.EqualValues(1, inputSelection.Length())

	name, ok := inputSelection.Attr("name")
	a.Equal("pub_key", name, "wrong name on input field")

	newKey := "@x7iOLUcq3o+sjGeAnipvWeGzfuYgrXl8L4LYlxIhwDc=.ed25519"
	addVals := url.Values{
		// just any key that looks valid
		"pub_key": []string{newKey},
	}
	rec := ts.Client.PostForm(addURL.String(), addVals)
	a.Equal(http.StatusFound, rec.Code)

	a.Equal(1, ts.AllowListDB.AddCallCount())
	_, added := ts.AllowListDB.AddArgsForCall(0)
	a.Equal(newKey, added.Ref())
}

func TestAllowList(t *testing.T) {
	ts := newSession(t)
	a := assert.New(t)

	lst := admindb.ListEntries{
		{ID: 1, PubKey: refs.FeedRef{ID: bytes.Repeat([]byte{0}, 32), Algo: "fake"}},
		{ID: 2, PubKey: refs.FeedRef{ID: bytes.Repeat([]byte("1312"), 8), Algo: "test"}},
		{ID: 3, PubKey: refs.FeedRef{ID: bytes.Repeat([]byte("acab"), 8), Algo: "true"}},
	}
	ts.AllowListDB.ListReturns(lst, nil)

	html, resp := ts.Client.GetHTML("/allow-list", nil)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	assertLocalized(t, html, []localizedElement{
		{"#welcome", "AdminAllowListWelcome"},
		{"title", "AdminAllowListTitle"},
		{"#allowListCount", "ListCountPlural"},
	})

	a.EqualValues(html.Find("#theList").Children().Length(), 3)

	lst = admindb.ListEntries{
		{ID: 666, PubKey: refs.FeedRef{ID: bytes.Repeat([]byte{1}, 32), Algo: "one"}},
	}
	ts.AllowListDB.ListReturns(lst, nil)

	html, resp = ts.Client.GetHTML("/allow-list", nil)
	a.Equal(http.StatusOK, resp.Code, "wrong HTTP status code")

	assertLocalized(t, html, []localizedElement{
		{"#welcome", "AdminAllowListWelcome"},
		{"title", "AdminAllowListTitle"},
		{"#allowListCount", "ListCountPlural"}, // TODO: should be singular - template func testing stub might have a qurik
	})

	elems := html.Find("#theList").Children()
	a.EqualValues(elems.Length(), 1)

	// check for link to remove confirm link
	link, yes := elems.ContentsFiltered("a").Attr("href")
	a.True(yes, "a-tag has href attribute")
	a.Equal("/allow-list/remove/confirm?id=666", link)
}
