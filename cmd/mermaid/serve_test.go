package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRenderHandler(t *testing.T) {
	Convey("Given the render HTTP handler", t, func() {
		srv := httptest.NewServer(renderHandler())
		defer srv.Close()

		Convey("When POSTing diagram source", func() {
			resp, err := http.Post(srv.URL, "text/plain", strings.NewReader("graph TD\nA-->B"))
			So(err, ShouldBeNil)
			defer func() { _ = resp.Body.Close() }()
			body := readAll(resp)

			Convey("Then it returns SVG with the right content type", func() {
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header.Get("Content-Type"), ShouldEqual, "image/svg+xml")
				So(body, ShouldStartWith, "<svg")
			})
		})

		Convey("When GETting with ?src= and ?theme=", func() {
			resp, err := http.Get(srv.URL + "/?src=" + "graph%20TD%0AA--%3EB" + "&theme=dark")
			So(err, ShouldBeNil)
			defer func() { _ = resp.Body.Close() }()
			body := readAll(resp)

			Convey("Then it renders with the requested theme", func() {
				So(resp.StatusCode, ShouldEqual, 200)
				So(body, ShouldContainSubstring, "#1e1e1e")
			})
		})

		Convey("When requesting PNG via ?format=png", func() {
			resp, err := http.Get(srv.URL + "/?src=" + "graph%20TD%0AA--%3EB" + "&format=png&scale=2")
			So(err, ShouldBeNil)
			defer func() { _ = resp.Body.Close() }()
			body := readAll(resp)

			Convey("Then it returns a PNG image", func() {
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header.Get("Content-Type"), ShouldEqual, "image/png")
				So(body[:4], ShouldEqual, "\x89PNG")
			})
		})

		Convey("When the source is invalid", func() {
			resp, err := http.Post(srv.URL, "text/plain", strings.NewReader("graph TD\nA[oops"))
			So(err, ShouldBeNil)
			defer func() { _ = resp.Body.Close() }()

			Convey("Then it responds 400", func() {
				So(resp.StatusCode, ShouldEqual, 400)
			})
		})
	})
}

func readAll(resp *http.Response) string {
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return sb.String()
}
