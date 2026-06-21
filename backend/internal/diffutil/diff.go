package diffutil

import (
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

type DiffLine struct {
	Type    string `json:"type"` // "equal", "insert", "delete"
	Content string `json:"content"`
	LineNum int    `json:"line_num"`
}

func ComputeUnifiedDiff(oldText, newText string) (string, error) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldText),
		B:        difflib.SplitLines(newText),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
	}
	return difflib.GetUnifiedDiffString(diff)
}

func ComputeLineDiff(oldText, newText string) []DiffLine {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	matcher := difflib.NewMatcher(oldLines, newLines)
	var result []DiffLine

	for _, group := range matcher.GetGroupedOpCodes(3) {
		for _, op := range group {
			switch op.Tag {
			case 'e':
				for i := op.I1; i < op.I2; i++ {
					result = append(result, DiffLine{
						Type:    "equal",
						Content: oldLines[i],
						LineNum: i + 1,
					})
				}
			case 'd':
				for i := op.I1; i < op.I2; i++ {
					result = append(result, DiffLine{
						Type:    "delete",
						Content: oldLines[i],
						LineNum: i + 1,
					})
				}
			case 'i':
				for j := op.J1; j < op.J2; j++ {
					result = append(result, DiffLine{
						Type:    "insert",
						Content: newLines[j],
						LineNum: j + 1,
					})
				}
			case 'r':
				for i := op.I1; i < op.I2; i++ {
					result = append(result, DiffLine{
						Type:    "delete",
						Content: oldLines[i],
						LineNum: i + 1,
					})
				}
				for j := op.J1; j < op.J2; j++ {
					result = append(result, DiffLine{
						Type:    "insert",
						Content: newLines[j],
						LineNum: j + 1,
					})
				}
			}
		}
	}

	return result
}
