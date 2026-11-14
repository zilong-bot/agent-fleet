package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Luolc/agent-fleet/internal/config"
)

// agyScreen is a screen agy 1.5.8 drew in a 150x45 terminal, from
// testdata/agy.
func agyScreen(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "agy", name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestAgyPromptBoxIsRecognisedAndItsDialogsAreNot(t *testing.T) {
	for name, want := range map[string]bool{
		"idle": true, "working": true, "done": true,
		"approval": false, "trust": false,
	} {
		if got := IsAgyInputBox(agyScreen(t, name)); got != want {
			t.Errorf("%s: IsAgyInputBox = %v, want %v", name, got, want)
		}
	}
	if IsAgyInputBox("") || IsAgyInputBox("> hello\n") {
		t.Error("a screen with no rules reads as the prompt box")
	}
}

func TestAgyFilterKeepsTheTranscriptAndDropsTheBoxAndSpinner(t *testing.T) {
	working := FilterAgyScreen(agyScreen(t, "working"))
	if strings.Contains(working, "Running command") || strings.Contains(working, "esc to cancel") {
		t.Errorf("the spinner or the footer is left:\n%s", working)
	}
	if !strings.Contains(working, "Run the shell command") {
		t.Errorf("the transcript is gone:\n%s", working)
	}
	if approval := agyScreen(t, "approval"); FilterAgyScreen(approval) != approval {
		t.Error("a dialog screen is not kept whole")
	}
}

func TestAgyGetsItsArgumentsAndTrustsOnlyItsDirectory(t *testing.T) {
	model, effort := "gemini-3.8-flash-low", "high"
	args := AgyArgs(&model, &effort)
	want := []string{"--dangerously-skip-permissions", "--model", model, "--effort", effort}
	if !slices.Equal(args, want) {
		t.Errorf("AgyArgs = %q, want %q", args, want)
	}
	if args := AgyArgs(nil, nil); !slices.Equal(args, []string{"--dangerously-skip-permissions"}) {
		t.Errorf("AgyArgs(nil, nil) = %q", args)
	}
	cwd := agyCwdArgs("/home/agent/wt/a \"b\"")
	if len(cwd) != 2 || cwd[0] != "--settings" || cwd[1] != `{"trustedWorkspaces":["/home/agent/wt/a \"b\""]}` {
		t.Errorf("agyCwdArgs = %q", cwd)
	}
	if CheckKind(KindAgy) != nil {
		t.Error("agy is refused")
	}
}

func TestTheScopeRuntimeChoosesTheKindAndOnlyClaudeGetsTheModels(t *testing.T) {
	if Runtime(nil) != KindClaude || Runtime(&config.Scope{}) != KindClaude {
		t.Error("no runtime is not Claude Code")
	}
	if Runtime(&config.Scope{Runtime: KindAgy}) != KindAgy {
		t.Error("runtime agy is not agy")
	}
	m := config.Model{Model: "sonnet", Effort: "high"}
	if model, effort := defaultsFor(KindAgy, m); model != nil || effort != nil {
		t.Error("agy gets Claude Code's model defaults")
	}
	if model, effort := defaultsFor(KindClaude, m); model == nil || *model != "sonnet" || effort == nil || *effort != "high" {
		t.Error("Claude Code does not get its model defaults")
	}
	if model, _ := defaultsFor(KindClaude, config.Model{}); model != nil {
		t.Error("an empty default is passed")
	}
}
