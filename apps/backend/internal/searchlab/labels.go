package searchlab

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Grades a result can get for a query. Unlabelled results have no grade.
const (
	GradeWrong   = 0 // not what the query asks for
	GradeLoose   = 1 // related, but not really it
	GradeGood    = 2 // a fitting result
	GradePerfect = 3 // exactly what the query wants
	// RelevantGrade is the lowest grade counting as a hit for precision.
	RelevantGrade = GradeGood
)

// labelColumns is the header of labels.tsv. Only query, url and grade are read
// back; the rest is there to make the file readable and editable by hand.
var labelColumns = []string{"query", "plugin", "rank", "grade", "title", "url"}

// Labels holds the hand-given grades per query and result key. Safe for
// concurrent use (the labelling page writes while it serves).
type Labels struct {
	mu     sync.Mutex
	grades map[string]map[string]int
}

// NewLabels returns an empty label set.
func NewLabels() *Labels {
	return &Labels{grades: map[string]map[string]int{}}
}

// Grade returns a result's grade for a query, false when unlabelled.
func (labels *Labels) Grade(queryID, key string) (int, bool) {
	labels.mu.Lock()
	defer labels.mu.Unlock()
	grade, found := labels.grades[queryID][key]
	return grade, found
}

// Set grades a result; a negative grade removes the label.
func (labels *Labels) Set(queryID, key string, grade int) {
	labels.mu.Lock()
	defer labels.mu.Unlock()
	if grade < 0 {
		delete(labels.grades[queryID], key)
		return
	}
	if labels.grades[queryID] == nil {
		labels.grades[queryID] = map[string]int{}
	}
	labels.grades[queryID][key] = grade
}

// LoadLabels reads labels.tsv; a missing file is an empty label set.
func LoadLabels(directory string) (*Labels, error) {
	labels := NewLabels()
	file, err := os.Open(filepath.Join(directory, LabelsFile))
	if errors.Is(err, fs.ErrNotExist) {
		return labels, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for line := 1; scanner.Scan(); line++ {
		if line == 1 {
			continue
		}
		if err := labels.readRow(scanner.Text()); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", LabelsFile, line, err)
		}
	}
	return labels, scanner.Err()
}

// readRow parses one labels.tsv row; an empty grade is an unlabelled result.
func (labels *Labels) readRow(row string) error {
	fields := strings.Split(row, "\t")
	if len(fields) != len(labelColumns) {
		return fmt.Errorf("want %d tab-separated columns, got %d", len(labelColumns), len(fields))
	}
	gradeText := strings.TrimSpace(fields[3])
	if gradeText == "" {
		return nil
	}
	grade, err := strconv.Atoi(gradeText)
	if err != nil || grade < GradeWrong || grade > GradePerfect {
		return fmt.Errorf("grade %q is not 0-3", gradeText)
	}
	labels.Set(fields[0], fields[5], grade)
	return nil
}

// SaveLabels writes labels.tsv with one row per snapshot result, graded or
// not, so the file always lists everything there is to label.
func SaveLabels(directory string, snapshot *Snapshot, labels *Labels) error {
	var buffer bytes.Buffer
	buffer.WriteString(strings.Join(labelColumns, "\t") + "\n")
	for _, list := range snapshot.Lists {
		for index, item := range list.Items {
			grade := ""
			if value, found := labels.Grade(list.QueryID, ItemKey(item)); found {
				grade = strconv.Itoa(value)
			}
			row := []string{list.QueryID, list.Plugin, strconv.Itoa(index + 1), grade, cleanCell(item.Title), ItemKey(item)}
			buffer.WriteString(strings.Join(row, "\t") + "\n")
		}
	}
	return writeAtomically(filepath.Join(directory, LabelsFile), buffer.Bytes())
}

// cleanCell keeps a value from breaking the TSV layout.
func cleanCell(value string) string {
	return strings.NewReplacer("\t", " ", "\n", " ", "\r", " ").Replace(value)
}
