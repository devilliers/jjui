package describe

import (
	"testing"

	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/operations"
	"github.com/idursun/jjui/test"
	"github.com/stretchr/testify/assert"
)

func TestEmbeddedHeight_UsesDynamicHeight(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.GetDescription("change")).SetOutput([]byte("this description should wrap onto multiple lines"))
	defer commandRunner.Verify()

	ctx := test.NewTestContext(commandRunner)
	op := NewOperation(ctx, &jj.Commit{ChangeId: "change", CommitId: "commit"})

	height := op.EmbeddedHeight(
		&jj.Commit{ChangeId: "change", CommitId: "commit"},
		operations.RenderOverDescription,
		12,
	)

	assert.Greater(t, height, 1)
}
