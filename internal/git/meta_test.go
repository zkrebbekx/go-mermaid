package git

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCommitMetaAndBranchOrder(t *testing.T) {
	Convey("Given a cherry-pick with an id", t, func() {
		d, err := Parse("gitGraph\ncommit\nbranch d\ncommit\ncheckout main\ncherry-pick id: \"abc\"")

		Convey("When parsing", func() {
			Convey("Then the id is kept instead of dropped", func() {
				So(err, ShouldBeNil)
				last := d.Commits[len(d.Commits)-1]
				So(last.ID, ShouldEqual, "abc")
			})
		})
	})

	Convey("Given a branch with an order hint", t, func() {
		d, err := Parse("gitGraph\ncommit\nbranch dev order: 2\ncommit")

		Convey("When parsing", func() {
			Convey("Then the branch is named without the hint", func() {
				So(err, ShouldBeNil)
				So(d.lane("dev"), ShouldBeGreaterThanOrEqualTo, 0)
				So(d.lane("dev order: 2"), ShouldEqual, -1)
			})
		})
	})
}
