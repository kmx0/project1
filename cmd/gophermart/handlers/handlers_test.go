package handlers

import (
	// "errors"

	"testing"

	"github.com/kmx0/project1/cmd/gophermart/storage"
	"github.com/kmx0/project1/internal/crypto"
	"github.com/kmx0/project1/internal/errors"
	"github.com/kmx0/project1/internal/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	x := 0
	// Only pass t into top-level Convey calls
	Convey("Тестируем сервер", t, func() {
		Convey("Тестируем публичную часть", func() {
			store := storage.NewMockStorage()

			id, err := store.RegisterUser(types.User{Login: "henry", Password: "blaqqq"})
			So(id, ShouldEqual, -1)
			So(err, ShouldBeError, errors.ErrStatusConflict.Error())

			id, err = store.RegisterUser(types.User{Login: "bla", Password: "bla"})
			So(id, ShouldEqual, 2)
			So(err, ShouldBeNil)

			id, err = store.RegisterUser(types.User{Login: "bla", Password: "bla"})
			So(id, ShouldEqual, -1)
			So(err, ShouldBeError, errors.ErrStatusConflict.Error())
			user := types.User{Login: "bla", Password: "bla", Cookie: crypto.CookieHash("", "", "bla")}

			err = store.WriteUserCookie(user, 2)
			So(err, ShouldBeNil)

			id, cookie, err := store.LoginUser(user)
			So(id, ShouldEqual, 2)
			So(cookie, ShouldEqual, user.Cookie)
			So(err, ShouldBeNil)
			// x++

			Convey("The value should be greater by one", func() {
				So(x, ShouldEqual, 0)
			})
		})
	})
}
