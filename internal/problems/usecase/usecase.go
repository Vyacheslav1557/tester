package usecase

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Vyacheslav1557/tester/internal/problems"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/microcosm-cc/bluemonday"
)

type UseCase struct {
	problemRepo  problems.Repository
	pandocClient pkg.PandocClient
	s3Repo       problems.S3Repository
	cacheDir     string
}

func NewUseCase(
	problemRepo problems.Repository,
	pandocClient pkg.PandocClient,
	s3Repo problems.S3Repository,
	cacheDir string,
) *UseCase {
	err := os.MkdirAll(path.Join(cacheDir, "archives"), 0755)
	if err != nil {
		panic(err) // FIXME: use logger
	}

	return &UseCase{
		problemRepo:  problemRepo,
		pandocClient: pandocClient,
		s3Repo:       s3Repo,
		cacheDir:     cacheDir,
	}
}

func (u *UseCase) CreateProblem(ctx context.Context, title string) (int32, error) {
	return u.problemRepo.CreateProblem(ctx, u.problemRepo.DB(), title)
}

func (u *UseCase) GetProblemById(ctx context.Context, id int32) (*models.Problem, error) {
	return u.problemRepo.GetProblemById(ctx, u.problemRepo.DB(), id)
}

func (u *UseCase) DownloadTestsArchive(ctx context.Context, id int32) (string, error) {
	rc, err := u.s3Repo.DownloadTestsFile(ctx, id)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	zipPath := path.Join(u.cacheDir, "archives", fmt.Sprintf("%d.zip", id))

	f, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, rc)
	if err != nil {
		return "", err
	}

	return zipPath, nil
}

func (u *UseCase) DeleteProblem(ctx context.Context, id int32) error {
	return u.problemRepo.DeleteProblem(ctx, u.problemRepo.DB(), id)
}

func (u *UseCase) ListProblems(ctx context.Context, filter models.ProblemsFilter) (*models.ProblemsList, error) {
	return u.problemRepo.ListProblems(ctx, u.problemRepo.DB(), filter)
}

func (u *UseCase) UpdateProblem(ctx context.Context, id int32, problemUpdate *models.ProblemUpdate) error {
	if isEmpty(*problemUpdate) {
		return pkg.Wrap(pkg.ErrBadInput, nil, "UpdateProblem", "empty problem update")
	}

	tx, err := u.problemRepo.BeginTx(ctx)
	if err != nil {
		return err
	}

	problem, err := u.problemRepo.GetProblemById(ctx, tx, id)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	statement := models.ProblemStatement{
		Legend:       problem.Legend,
		InputFormat:  problem.InputFormat,
		OutputFormat: problem.OutputFormat,
		Notes:        problem.Notes,
		Scoring:      problem.Scoring,
	}

	if problemUpdate.Legend != nil {
		statement.Legend = *problemUpdate.Legend
	}
	if problemUpdate.InputFormat != nil {
		statement.InputFormat = *problemUpdate.InputFormat
	}
	if problemUpdate.OutputFormat != nil {
		statement.OutputFormat = *problemUpdate.OutputFormat
	}
	if problemUpdate.Notes != nil {
		statement.Notes = *problemUpdate.Notes
	}
	if problemUpdate.Scoring != nil {
		statement.Scoring = *problemUpdate.Scoring
	}

	builtStatement, err := build(ctx, u.pandocClient, trimSpaces(statement))
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	problemUpdate.LegendHtml = &builtStatement.LegendHtml
	problemUpdate.InputFormatHtml = &builtStatement.InputFormatHtml
	problemUpdate.OutputFormatHtml = &builtStatement.OutputFormatHtml
	problemUpdate.NotesHtml = &builtStatement.NotesHtml
	problemUpdate.ScoringHtml = &builtStatement.ScoringHtml

	err = u.problemRepo.UpdateProblem(ctx, tx, id, problemUpdate)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type ProblemProperties struct {
	Title string `json:"name"`

	TimeLimit   int64 `json:"timeLimit"`
	MemoryLimit int64 `json:"memoryLimit"`

	Legend       *string `json:"legend"`
	Scoring      *string `json:"scoring"`
	Notes        *string `json:"notes"`
	OutputFormat *string `json:"output"`
	InputFormat  *string `json:"input"`

	Meta *models.Meta

	//Tutorial    *string      `json:"tutorial"`
	//InputFile   string       `json:"inputFile"`
	//OutputFile  string       `json:"outputFile"`
	//AuthorName  string       `json:"authorName"`
	//Language    string       `json:"language"`
	//SampleTests []SampleTest `json:"sampleTests"`
	//Interaction *string      `json:"interaction"`
	//AuthorLogin string       `json:"authorLogin"`
}

func (u *UseCase) UploadProblem(ctx context.Context, id int32, r io.ReaderAt, size int64) error {
	const op = "UseCase.UploadProblem"

	// Initialize zip reader
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to open zip")
	}

	// Process zip contents
	properties, testsBuffer, err := processZipContents(ctx, zipReader)
	if err != nil {
		return err
	}

	// Update problem properties
	problemUpdate := &models.ProblemUpdate{
		Title: &properties.Title,

		TimeLimit:   int32p(int32(properties.TimeLimit)),
		MemoryLimit: int32p(int32(properties.MemoryLimit)),

		Legend:       properties.Legend,
		InputFormat:  properties.InputFormat,
		OutputFormat: properties.OutputFormat,
		Notes:        properties.Notes,
		Scoring:      properties.Scoring,

		Meta: properties.Meta,
	}

	if err := u.UpdateProblem(ctx, id, problemUpdate); err != nil {
		return err
	}

	// Upload tests to S3
	if _, err := u.s3Repo.UploadTestsFile(ctx, id, bytes.NewReader(testsBuffer.Bytes())); err != nil {
		return err
	}

	return nil
}

func processZipContents(_ context.Context, zipReader *zip.Reader) (*ProblemProperties, *bytes.Buffer, error) {
	const op = "processZipContents"

	const locale = "russian"

	testsBuffer := &bytes.Buffer{}
	testsArchive := zip.NewWriter(testsBuffer)

	var properties *ProblemProperties
	var meta models.Meta
	testInputs := make(map[string]bool)
	testOutputs := make(map[string]bool)

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() || isInvalidTestFile(file.Name) {
			continue
		}

		if file.Name == fmt.Sprintf("statements/%s/problem-properties.json", locale) {
			var err error
			properties, err = readProperties(file)
			if err != nil {
				return nil, nil, pkg.Wrap(pkg.ErrBadInput, err,
					op, "failed to read problem-properties.json")
			}
			continue
		}

		if strings.HasPrefix(file.Name, "tests/") && filepath.Dir(file.Name) == "tests" {
			fileName := filepath.Base(file.Name)
			if strings.HasSuffix(fileName, ".a") {
				testOutputs[strings.TrimSuffix(fileName, ".a")] = true
			} else {
				testInputs[fileName] = true
			}

			if err := copyTestFile(file, testsArchive); err != nil {
				return nil, nil, pkg.Wrap(pkg.ErrBadInput, err, op, "failed to copy test file")
			}
		}
	}

	if properties == nil {
		return nil, nil, pkg.Wrap(pkg.ErrBadInput, nil, op, "problem-properties.json not found")
	}

	if err := validateTests(testInputs, testOutputs); err != nil {
		return nil, nil, err
	}

	names := make([]string, 0, len(testInputs))
	for input := range testInputs {
		names = append(names, input)
	}
	meta.Names = names
	meta.Count = len(meta.Names)
	properties.MemoryLimit /= 1024 * 1024 // Convert bytes to MB
	properties.Meta = &meta

	if err := testsArchive.Close(); err != nil {
		return nil, nil, err
	}

	return properties, testsBuffer, nil
}

func isInvalidTestFile(name string) bool {
	fileName := filepath.Base(name)
	return fileName == "" || strings.HasPrefix(fileName, ".")
}

func copyTestFile(src *zip.File, dst *zip.Writer) error {
	srcReader, err := src.Open()
	if err != nil {
		return fmt.Errorf("failed to open test file: %w", err)
	}
	defer srcReader.Close()

	dstWriter, err := dst.Create(src.Name)
	if err != nil {
		return fmt.Errorf("failed to create test file in archive: %w", err)
	}

	if _, err := io.Copy(dstWriter, srcReader); err != nil {
		return fmt.Errorf("failed to copy test file: %w", err)
	}

	return nil
}

func validateTests(inputs, outputs map[string]bool) error {
	for input := range inputs {
		if !outputs[input] {
			return pkg.Wrap(pkg.ErrBadInput, nil, "validateTests",
				"missing output file for test input "+input)
		}
	}
	for output := range outputs {
		if !inputs[output] {
			return pkg.Wrap(pkg.ErrBadInput, nil, "validateTests",
				"missing input file for test output "+output)
		}
	}
	return nil
}

func readProperties(f *zip.File) (*ProblemProperties, error) {
	file, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var properties ProblemProperties
	if err := json.NewDecoder(file).Decode(&properties); err != nil {
		return nil, err
	}

	return &properties, nil
}

func isEmpty(p models.ProblemUpdate) bool {
	return p.Title == nil &&
		p.Legend == nil &&
		p.InputFormat == nil &&
		p.OutputFormat == nil &&
		p.Notes == nil &&
		p.Scoring == nil &&
		p.MemoryLimit == nil &&
		p.TimeLimit == nil
}

func wrap(s string) string {
	return fmt.Sprintf("\\begin{document}\n%s\n\\end{document}\n", s)
}

func trimSpaces(statement models.ProblemStatement) models.ProblemStatement {
	return models.ProblemStatement{
		Legend:       strings.TrimSpace(statement.Legend),
		InputFormat:  strings.TrimSpace(statement.InputFormat),
		OutputFormat: strings.TrimSpace(statement.OutputFormat),
		Notes:        strings.TrimSpace(statement.Notes),
		Scoring:      strings.TrimSpace(statement.Scoring),
	}
}

func sanitize(statement models.Html5ProblemStatement) models.Html5ProblemStatement {
	p := bluemonday.UGCPolicy()

	p.AllowAttrs("class").Globally()
	p.AllowAttrs("style").Globally()
	p.AllowStyles("text-align").MatchingEnum("center", "left", "right").Globally()
	p.AllowStyles("display").MatchingEnum("block", "inline", "inline-block").Globally()

	p.AllowStandardURLs()
	p.AllowAttrs("cite").OnElements("blockquote", "q")
	p.AllowAttrs("href").OnElements("a", "area")
	p.AllowAttrs("src").OnElements("img")

	if statement.LegendHtml != "" {
		statement.LegendHtml = p.Sanitize(statement.LegendHtml)
	}
	if statement.InputFormatHtml != "" {
		statement.InputFormatHtml = p.Sanitize(statement.InputFormatHtml)
	}
	if statement.OutputFormatHtml != "" {
		statement.OutputFormatHtml = p.Sanitize(statement.OutputFormatHtml)
	}
	if statement.NotesHtml != "" {
		statement.NotesHtml = p.Sanitize(statement.NotesHtml)
	}
	if statement.ScoringHtml != "" {
		statement.ScoringHtml = p.Sanitize(statement.ScoringHtml)
	}

	return statement
}

func build(ctx context.Context, pandocClient pkg.PandocClient, p models.ProblemStatement) (models.Html5ProblemStatement, error) {
	p = trimSpaces(p)

	latex := models.ProblemStatement{}

	if p.Legend != "" {
		latex.Legend = wrap(p.Legend)
	}
	if p.InputFormat != "" {
		latex.InputFormat = wrap(p.InputFormat)
	}
	if p.OutputFormat != "" {
		latex.OutputFormat = wrap(p.OutputFormat)
	}
	if p.Notes != "" {
		latex.Notes = wrap(p.Notes)
	}
	if p.Scoring != "" {
		latex.Scoring = wrap(p.Scoring)
	}

	req := []string{
		latex.Legend,
		latex.InputFormat,
		latex.OutputFormat,
		latex.Notes,
		latex.Scoring,
	}

	res, err := pandocClient.BatchConvertLatexToHtml5(ctx, req)
	if err != nil {
		return models.Html5ProblemStatement{}, err
	}

	if len(res) != len(req) {
		return models.Html5ProblemStatement{}, fmt.Errorf("wrong number of fieilds returned: %d", len(res))
	}

	sanitizedStatement := sanitize(models.Html5ProblemStatement{
		LegendHtml:       res[0],
		InputFormatHtml:  res[1],
		OutputFormatHtml: res[2],
		NotesHtml:        res[3],
		ScoringHtml:      res[4],
	})

	return sanitizedStatement, nil
}

func int32p(v int32) *int32 {
	return &v
}
