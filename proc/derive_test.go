package proc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Derive(t *testing.T) {
	derive, err := Match(`#[ident(k1={k2="v2",k3="v3"})]`)
	require.NoError(t, err)
	require.False(t, derive.Headless())

	object := derive.Attribute()
	vt := object["k1"]
	require.NotNil(t, vt)

	val, ok := vt.(Map)
	require.True(t, ok)

	obj := val.EntryMap()
	require.Equal(t, String{Value: "v2"}, obj["k2"])
	require.Equal(t, String{Value: "v3"}, obj["k3"])
}

const testComment = `// #[ident(k1="v1")]
// #[ident(k2="v2")]
// #[enum]
// 我是个注释`

func Test_Derives(t *testing.T) {
	t.Run("", func(t *testing.T) {
		line := NewCommentLines(testComment)
		derives1, line1 := line.Derives()
		require.Equal(t, 3, len(derives1))
		require.Equal(t, 1, len(line1))
		require.True(t, Derives(derives1).ContainHeadless("enum"))
		require.False(t, Derives(derives1).ContainHeadless("ident"))
		require.Equal(t, 2, len(Derives(derives1).Find("ident")))
		require.Equal(t, []ValueType{String{Value: "v1"}}, Derives(derives1).FindValue("ident", "k1"))

		derives2, line2 := line.FindDerives("ident")
		require.Equal(t, 2, len(derives2))
		require.Equal(t, 1, len(line2))

		require.Equal(t, testComment, line.String())

		line = line.Append("我是个注释2")
		require.Equal(t, `#[ident(k1="v1")], #[ident(k2="v2")], #[enum], 我是个注释, 我是个注释2`, line.LineString())
	})

	t.Run("", func(t *testing.T) {
		line := NewCommentLines(`#[ident(k1="v1")]`)

		require.Equal(t, `// #[ident(k1="v1")]`, line.String())
		require.Equal(t, `#[ident(k1="v1")]`, line.LineString())
	})

	t.Run("empty", func(t *testing.T) {
		line := NewCommentLines("")

		require.Equal(t, "", line.String())
		require.Equal(t, "", line.LineString())
	})
}
