package envsync

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Paths struct {
	ConfigPath string
	Files      []string
}

type Entry struct {
	Name       string
	Default    string
	HasDefault bool
}

type Schema struct {
	Entries []Entry
}

type FileIssue struct {
	Path    string
	Missing []string
	Stale   []string
}

type CheckError struct {
	Issues []FileIssue
}

type FixReport struct {
	Files []FileChange
}

type FileChange struct {
	Path      string
	Created   bool
	Added     []string
	Commented []string
}

func (e *CheckError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("envsync check failed")

	for _, issue := range e.Issues {
		builder.WriteString("\n- ")
		builder.WriteString(issue.Path)

		if len(issue.Missing) > 0 {
			builder.WriteString(" missing: ")
			builder.WriteString(strings.Join(issue.Missing, ", "))
		}

		if len(issue.Stale) > 0 {
			builder.WriteString(" stale: ")
			builder.WriteString(strings.Join(issue.Stale, ", "))
		}
	}

	return builder.String()
}

func Fix(paths Paths) (*FixReport, error) {
	schema, err := LoadSchema(paths.ConfigPath)
	if err != nil {
		return nil, err
	}

	report := &FixReport{}

	for _, filePath := range uniquePaths(paths.Files) {
		change, err := syncFile(filePath, schema)
		if err != nil {
			return nil, err
		}

		if change.Path != "" {
			report.Files = append(report.Files, change)
		}
	}

	return report, nil
}

func Check(paths Paths) error {
	schema, err := LoadSchema(paths.ConfigPath)
	if err != nil {
		return err
	}

	keySet := schemaKeySet(schema)
	var issues []FileIssue

	for _, filePath := range uniquePaths(paths.Files) {
		issue, err := compareFile(filePath, schema, keySet)
		if err != nil {
			return err
		}

		if len(issue.Missing) > 0 || len(issue.Stale) > 0 {
			issues = append(issues, issue)
		}
	}

	if len(issues) > 0 {
		return &CheckError{Issues: issues}
	}

	return nil
}

func LoadSchema(configPath string) (Schema, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, configPath, nil, parser.ParseComments)
	if err != nil {
		return Schema{}, err
	}

	structs := make(map[string]*ast.StructType)

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structs[typeSpec.Name.Name] = structType
		}
	}

	entries, err := collectEntries("Config", structs, map[string]bool{})
	if err != nil {
		return Schema{}, err
	}

	seen := make(map[string]struct{}, len(entries))
	ordered := make([]Entry, 0, len(entries))

	for _, entry := range entries {
		if _, exists := seen[entry.Name]; exists {
			return Schema{}, fmt.Errorf("duplicate env key %q in %s", entry.Name, configPath)
		}

		seen[entry.Name] = struct{}{}
		ordered = append(ordered, entry)
	}

	return Schema{Entries: ordered}, nil
}

func collectEntries(structName string, structs map[string]*ast.StructType, visiting map[string]bool) ([]Entry, error) {
	if visiting[structName] {
		return nil, fmt.Errorf("cyclic struct reference detected at %q", structName)
	}

	structType, ok := structs[structName]
	if !ok {
		return nil, fmt.Errorf("struct %q not found in config", structName)
	}

	visiting[structName] = true
	defer delete(visiting, structName)

	return collectEntriesFromFields(structType.Fields.List, structs, visiting)
}

func collectEntriesFromFields(fields []*ast.Field, structs map[string]*ast.StructType, visiting map[string]bool) ([]Entry, error) {
	var entries []Entry

	for _, field := range fields {
		tag := parseTag(field)

		if len(field.Names) > 0 {
			if envName := tag.Get("env"); envName != "" {
				entry := Entry{Name: envName}
				if defaultValue, ok := tag.Lookup("envDefault"); ok {
					entry.Default = defaultValue
					entry.HasDefault = true
				}

				entries = append(entries, entry)
				continue
			}

			if nestedStruct, ok := field.Type.(*ast.StructType); ok {
				subEntries, err := collectEntriesFromFields(nestedStruct.Fields.List, structs, visiting)
				if err != nil {
					return nil, err
				}

				entries = append(entries, subEntries...)
				continue
			}

			if referencedName, ok := embeddedNameFromExpr(field.Type); ok {
				if _, exists := structs[referencedName]; exists {
					subEntries, err := collectEntries(referencedName, structs, visiting)
					if err != nil {
						return nil, err
					}

					entries = append(entries, subEntries...)
				}
			}

			continue
		}

		embeddedName, ok := embeddedNameFromExpr(field.Type)
		if !ok {
			continue
		}

		if _, exists := structs[embeddedName]; !exists {
			continue
		}

		subEntries, err := collectEntries(embeddedName, structs, visiting)
		if err != nil {
			return nil, err
		}

		entries = append(entries, subEntries...)
	}

	return entries, nil
}

func compareFile(filePath string, schema Schema, keySet map[string]struct{}) (FileIssue, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			missing := make([]string, 0, len(schema.Entries))
			for _, entry := range schema.Entries {
				missing = append(missing, entry.Name)
			}

			return FileIssue{Path: filePath, Missing: missing}, nil
		}

		return FileIssue{}, err
	}

	existing := parseEnvEntries(string(content))
	issue := FileIssue{Path: filePath}

	for _, entry := range schema.Entries {
		if _, ok := existing.active[entry.Name]; !ok {
			issue.Missing = append(issue.Missing, entry.Name)
		}
	}

	for key := range existing.active {
		if _, ok := keySet[key]; !ok {
			issue.Stale = append(issue.Stale, key)
		}
	}

	return issue, nil
}

func syncFile(filePath string, schema Schema) (FileChange, error) {
	content, err := os.ReadFile(filePath)
	created := false
	if err != nil && !os.IsNotExist(err) {
		return FileChange{}, err
	}

	if os.IsNotExist(err) {
		created = true
	}

	updated, changed, change := reconcileFile(string(content), schema, filePath, created)
	if !changed && !created {
		return FileChange{}, nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return FileChange{}, err
	}

	if err := os.WriteFile(filePath, []byte(updated), 0o644); err != nil {
		return FileChange{}, err
	}

	return change, nil
}

func reconcileFile(content string, schema Schema, filePath string, created bool) (string, bool, FileChange) {
	existing := parseEnvEntries(content)
	keySet := schemaKeySet(schema)
	lines := make([]string, 0, len(existing.lines)+len(schema.Entries))
	changed := false
	change := FileChange{Path: filePath, Created: created}

	if created {
		for _, entry := range schema.Entries {
			change.Added = append(change.Added, entry.Name)
			lines = append(lines, renderEntry(entry))
		}

		updated := strings.Join(lines, detectNewline(content))
		if updated != "" {
			updated += detectNewline(content)
		}

		return updated, true, change
	}

	for _, line := range existing.lines {
		switch {
		case line.kind != envLineActive:
			lines = append(lines, line.raw)
		case line.key != "" && hasKey(keySet, line.key):
			lines = append(lines, line.raw)
		default:
			lines = append(lines, commentLine(line.raw))
			if line.key != "" {
				change.Commented = append(change.Commented, line.key)
			}
			changed = true
		}
	}

	for _, entry := range schema.Entries {
		if _, ok := existing.active[entry.Name]; ok {
			continue
		}

		lines = append(lines, renderEntry(entry))
		change.Added = append(change.Added, entry.Name)
		changed = true
	}

	if len(lines) == 0 {
		return content, false, change
	}

	newline := detectNewline(content)
	updated := strings.Join(lines, newline)
	if updated != "" {
		updated += newline
	}

	return updated, changed, change
}

func renderEntry(entry Entry) string {
	if entry.HasDefault {
		return fmt.Sprintf("%s=%s", entry.Name, entry.Default)
	}

	return fmt.Sprintf("%s=", entry.Name)
}

type envLineKind int

const (
	envLineComment envLineKind = iota
	envLineActive
)

type parsedEnv struct {
	active map[string]struct{}
	lines  []parsedLine
}

type parsedLine struct {
	raw  string
	key  string
	kind envLineKind
}

func parseEnvEntries(content string) parsedEnv {
	parsed := parsedEnv{
		active: make(map[string]struct{}),
		lines:  make([]parsedLine, 0),
	}

	for _, rawLine := range strings.Split(content, "\n") {
		trimmedLine := strings.TrimSuffix(rawLine, "\r")
		line := strings.TrimSpace(trimmedLine)
		if line == "" {
			parsed.lines = append(parsed.lines, parsedLine{raw: trimmedLine, kind: envLineComment})
			continue
		}

		if strings.HasPrefix(line, "#") {
			parsed.lines = append(parsed.lines, parsedLine{raw: trimmedLine, kind: envLineComment})
			continue
		}

		key, _, ok := strings.Cut(line, "=")
		if !ok {
			parsed.lines = append(parsed.lines, parsedLine{raw: trimmedLine, kind: envLineComment})
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			parsed.lines = append(parsed.lines, parsedLine{raw: trimmedLine, kind: envLineComment})
			continue
		}

		parsed.active[key] = struct{}{}
		parsed.lines = append(parsed.lines, parsedLine{raw: trimmedLine, key: key, kind: envLineActive})
	}

	return parsed
}

func schemaKeySet(schema Schema) map[string]struct{} {
	keys := make(map[string]struct{}, len(schema.Entries))
	for _, entry := range schema.Entries {
		keys[entry.Name] = struct{}{}
	}

	return keys
}

func hasKey(keys map[string]struct{}, key string) bool {
	_, ok := keys[key]
	return ok
}

func commentLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}

	return "# " + line
}

func parseTag(field *ast.Field) reflect.StructTag {
	if field == nil || field.Tag == nil {
		return ""
	}

	return reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
}

func embeddedNameFromExpr(expr ast.Expr) (string, bool) {
	switch typeExpr := expr.(type) {
	case *ast.Ident:
		return typeExpr.Name, true
	case *ast.StarExpr:
		return embeddedNameFromExpr(typeExpr.X)
	case *ast.SelectorExpr:
		if typeExpr.Sel != nil {
			return typeExpr.Sel.Name, true
		}
	}

	return "", false
}

func detectNewline(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}

	return "\n"
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))

	for _, filePath := range paths {
		if filePath == "" {
			continue
		}

		if _, ok := seen[filePath]; ok {
			continue
		}

		seen[filePath] = struct{}{}
		result = append(result, filePath)
	}

	return result
}
