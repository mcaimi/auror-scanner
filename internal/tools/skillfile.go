package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// SkillFileInput is the input schema for the skill file reading tool.
type SkillFileInput struct {
	// SkillName is the name of the skill (its directory name under the skills path).
	SkillName string `json:"skill_name"`
	// FilePath is the relative path to the file within the skill directory
	// (e.g. "references/PROCEDURE.md" or "scripts/entropy_scan.py").
	FilePath string `json:"file_path"`
}

// SkillFileOutput holds the content of the requested skill file.
type SkillFileOutput struct {
	Content string `json:"content"`
	Path    string `json:"path"`
}

// RegisterSkillFileTool defines and registers a genkit tool that reads files
// from a skill's directory tree. The model calls this tool to lazily load
// reference documents, scripts, or other assets referenced by SKILL.md.
//
// skillsPath is the base directory containing skill subdirectories (e.g. "skill").
func RegisterSkillFileTool(g *genkit.Genkit, skillsPath string) *ai.ToolDef[SkillFileInput, SkillFileOutput] {
	absSkillsPath, err := filepath.Abs(skillsPath)
	if err != nil {
		absSkillsPath = skillsPath
	}

	return genkit.DefineTool(
		g,
		"read_skill_file",
		"Read a file from a skill's directory. Use this to load reference documents, "+
			"scripts, or other assets that are referenced by relative path in SKILL.md. "+
			"Provide the skill name (directory name) and the relative file path "+
			"(e.g. skill_name='pkgbuild-security-assessment', file_path='references/PROCEDURE.md').",
		func(tctx *ai.ToolContext, input SkillFileInput) (SkillFileOutput, error) {
			if input.SkillName == "" || input.FilePath == "" {
				return SkillFileOutput{}, fmt.Errorf("read_skill_file: skill_name and file_path are required")
			}

			target := filepath.Join(absSkillsPath, input.SkillName, input.FilePath)
			resolved, err := filepath.Abs(target)
			if err != nil {
				return SkillFileOutput{}, fmt.Errorf("read_skill_file: invalid path: %w", err)
			}

			skillDir := filepath.Join(absSkillsPath, input.SkillName) + string(filepath.Separator)
			if !strings.HasPrefix(resolved, skillDir) && resolved != strings.TrimSuffix(skillDir, string(filepath.Separator)) {
				return SkillFileOutput{}, fmt.Errorf("read_skill_file: path escapes skill directory")
			}

			data, err := os.ReadFile(resolved)
			if err != nil {
				return SkillFileOutput{}, fmt.Errorf("read_skill_file: %w", err)
			}

			return SkillFileOutput{
				Content: string(data),
				Path:    resolved,
			}, nil
		},
	)
}
