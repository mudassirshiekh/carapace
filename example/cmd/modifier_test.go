package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/pkg/sandbox"
	"github.com/rsteube/carapace/pkg/style"
)

func TestBatch(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("modifier", "--batch", "").
			Expect(carapace.ActionValuesDescribed(
				"A", "description of A",
				"B", "description of second B",
				"C", "description of second C",
				"D", "description of D",
			).
				Usage("Batch()"))
	})
}

func TestCache(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		cached := s.Run("modifier", "--cache", "").Output()
		time.Sleep(1 * time.Second)
		s.Run("modifier", "--cache", "").
			Expect(cached.
				Usage("Cache()"))

		s.ClearCache()
		s.Run("modifier", "--cache", "").
			ExpectNot(cached.
				Usage("Cache()"))

	})
}

func TestFilter(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("modifier", "--filter", "").
			Expect(carapace.ActionValuesDescribed(
				"1", "one",
				"3", "three",
			).Usage("Filter()"))
	})
}

func TestRetain(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("modifier", "--retain", "").
			Expect(carapace.ActionValuesDescribed(
				"2", "two",
				"4", "four",
			).Usage("Retain()"))
	})
}

func TestShift(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("modifier", "one", "--shift", "").
			Expect(carapace.ActionMessage(`[]string{}`).Usage("Shift()"))

		s.Run("modifier", "one", "two", "three", "--shift", "").
			Expect(carapace.ActionMessage(`[]string{"two", "three"}`).Usage("Shift()"))
	})
}

func TestTimeout(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("modifier", "--timeout", "").
			Expect(carapace.ActionMessage("timeout exceeded").
				Usage("Timeout()"))
	})
}

func TestUsage(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("modifier", "--usage", "").
			Expect(carapace.ActionValues().
				Usage("explicit usage"))
	})
}

func TestChdir(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return carapace.ActionFiles().Chdir("subdir")
	})(func(s *sandbox.Sandbox) {
		s.Files("subdir/file1.txt", "")

		s.Run("").Expect(
			carapace.ActionValues("file1.txt").
				StyleF(func(s string, sc style.Context) string {
					return style.ForPath("subdir/file1.txt", sc)
				}).
				NoSpace('/').
				Tag("files"))
	})
}

func TestMultiParts(t *testing.T) {
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("modifier", "--multiparts", "").
			Expect(carapace.ActionValues("dir/").
				NoSpace('/').
				Usage("MultiParts()"))

		s.Run("modifier", "--multiparts", "dir/").
			Expect(carapace.ActionValues("subdir1/", "subdir2/").
				Prefix("dir/").
				NoSpace('/').
				Usage("MultiParts()"))

		s.Run("modifier", "--multiparts", "dir/subdir1/").
			Expect(carapace.ActionValues("fileA.txt", "fileB.txt").
				Prefix("dir/subdir1/").
				NoSpace('/').
				Usage("MultiParts()"))

		s.Run("modifier", "--multiparts", "dir/subdir2/").
			Expect(carapace.ActionValues("fileC.txt").
				Prefix("dir/subdir2/").
				NoSpace('/').
				Usage("MultiParts()"))
	})
}

func TestPrefix(t *testing.T) {
	os.Unsetenv("LS_COLORS")
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Files("subdir/file1.txt", "")

		s.Run("modifier", "--prefix", "").
			Expect(carapace.ActionValues("subdir/").
				StyleF(style.ForPath).
				Prefix("file://").
				NoSpace('/').
				Usage("Prefix()").
				Tag("files"))

		s.Run("modifier", "--prefix", "file").
			Expect(carapace.ActionValues("subdir/").
				StyleF(style.ForPath).
				Prefix("file://").
				NoSpace('/').
				Usage("Prefix()").
				Tag("files"))

		s.Run("modifier", "--prefix", "file://subdir/f").
			Expect(carapace.ActionValues("file1.txt").
				StyleF(style.ForPath).
				Prefix("file://subdir/").
				NoSpace('/').
				Usage("Prefix()").
				Tag("files"))
	})
}

func TestSplit(t *testing.T) {
	os.Unsetenv("LS_COLORS")
	sandbox.Package(t, "github.com/rsteube/carapace/example")(func(s *sandbox.Sandbox) {
		s.Files("subdir/file1.txt", "")

		s.Run("modifier", "--split", "").
			Expect(carapace.ActionValues(
				"pos1",
				"positional1",
			).NoSpace('*').
				Suffix(" ").
				Usage("Split()"))

		s.Run("modifier", "--split", "pos1 ").
			Expect(carapace.ActionValues(
				"subdir/",
			).StyleF(style.ForPathExt).
				Prefix("pos1 ").
				NoSpace('*').
				Usage("Split()").
				Tag("files"))

		s.Run("modifier", "--split", "pos1 --").
			Expect(carapace.ActionStyledValuesDescribed(
				"--bool", "bool flag", style.Default,
				"--string", "string flag", style.Blue,
			).Prefix("pos1 ").
				Suffix(" ").
				NoSpace('*').
				Usage("Split()").
				Tag("flags"))

		s.Run("modifier", "--split", "pos1 --bool=").
			Expect(carapace.ActionStyledValues(
				"true", style.Green,
				"false", style.Red,
			).Prefix("pos1 --bool=").
				Suffix(" ").
				NoSpace('*').
				Usage("bool flag"))

		s.Run("modifier", "--split", "pos1 \"--bool=").
			Expect(carapace.ActionStyledValues(
				"true", style.Green,
				"false", style.Red,
			).Prefix("pos1 \"--bool=").
				Suffix("\" ").
				NoSpace('*').
				Usage("bool flag"))

		s.Run("modifier", "--split", "pos1 '--bool=").
			Expect(carapace.ActionStyledValues(
				"true", style.Green,
				"false", style.Red,
			).Prefix("pos1 '--bool=").
				Suffix("' ").
				NoSpace('*').
				Usage("bool flag"))

		t.Skip("skipping test that don't work yet") // TODO these need to work
		s.Run("modifier", "--split", "pos1 \"").
			Expect(carapace.ActionValues(
				"subdir/",
			).StyleF(style.ForPathExt).
				Prefix("pos1 \"").
				Suffix("\"").
				NoSpace('*').
				Usage("Split()").
				Tag("files"))

		s.Run("modifier", "--split", "pos1 '").
			Expect(carapace.ActionValues(
				"subdir/",
			).StyleF(style.ForPathExt).
				Prefix("pos1 '").
				Suffix("'").
				NoSpace('*').
				Usage("Split()").
				Tag("files"))
	})
}
