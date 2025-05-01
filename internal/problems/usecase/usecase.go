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
	"strings"

	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/microcosm-cc/bluemonday"
)

type UseCase struct {
	problemRepo  problems.Repository
	pandocClient pkg.PandocClient
}

func NewUseCase(
	problemRepo problems.Repository,
	pandocClient pkg.PandocClient,
) *UseCase {
	return &UseCase{
		problemRepo:  problemRepo,
		pandocClient: pandocClient,
	}
}

func (u *UseCase) CreateProblem(ctx context.Context, title string) (int32, error) {
	return u.problemRepo.CreateProblem(ctx, u.problemRepo.DB(), title)
}

func (u *UseCase) GetProblemById(ctx context.Context, id int32) (*models.Problem, error) {
	return u.problemRepo.GetProblemById(ctx, u.problemRepo.DB(), id)
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

	if builtStatement.LegendHtml != problem.LegendHtml {
		problemUpdate.LegendHtml = &builtStatement.LegendHtml
	}
	if builtStatement.InputFormatHtml != problem.InputFormatHtml {
		problemUpdate.InputFormatHtml = &builtStatement.InputFormatHtml
	}
	if builtStatement.OutputFormatHtml != problem.OutputFormatHtml {
		problemUpdate.OutputFormatHtml = &builtStatement.OutputFormatHtml
	}
	if builtStatement.NotesHtml != problem.NotesHtml {
		problemUpdate.NotesHtml = &builtStatement.NotesHtml
	}
	if builtStatement.ScoringHtml != problem.ScoringHtml {
		problemUpdate.ScoringHtml = &builtStatement.ScoringHtml
	}

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
	Title       string `json:"name"`
	TimeLimit   int32  `json:"timeLimit"`
	MemoryLimit int32  `json:"memoryLimit"`
}

func (u *UseCase) UploadProblem(ctx context.Context, id int32, data []byte) error {

	locale := "russian"
	defaultLocale := "english"
	var localeProblem, defaultProblem string
	var localeProperties, defaultProperties ProblemProperties

	r := bytes.NewReader(data)
	rc, err := zip.NewReader(r, int64(r.Len()))
	if err != nil {
		return err
	}

	testsZipBuf := new(bytes.Buffer)
	w := zip.NewWriter(testsZipBuf)

	for _, f := range rc.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if f.Name == fmt.Sprintf("statements/%s/problem.tex", locale) {
			localeProblem, err = readProblem(f)
			if err != nil {
				return err
			}
		}
		if f.Name == fmt.Sprintf("statements/%s/problem.tex", defaultLocale) {
			defaultProblem, err = readProblem(f)
			if err != nil {
				return err
			}
		}
		if f.Name == fmt.Sprintf("statements/%s/problem-properties.json", locale) {
			localeProperties, err = readProperties(f)
			if err != nil {
				return err
			}
		}
		if f.Name == fmt.Sprintf("statements/%s/problem-properties.json", defaultLocale) {
			defaultProperties, err = readProperties(f)
			if err != nil {
				return err
			}
		}
		if strings.HasPrefix(f.Name, "tests/") {
			if err := w.Copy(f); err != nil {
				return err
			}
		}
	}

	if err := w.Close(); err != nil {
		return err
	}
	// testsZipBuf contains test files; this is for s3

	localeProperties.MemoryLimit /= 1024 * 1024
	defaultProperties.MemoryLimit /= 1024 * 1024

	problemUpdate := &models.ProblemUpdate{}
	if localeProblem != "" {
		problemUpdate.Legend = &localeProblem
		problemUpdate.Title = &localeProperties.Title
		problemUpdate.TimeLimit = &localeProperties.TimeLimit
		problemUpdate.MemoryLimit = &localeProperties.MemoryLimit
	} else {
		problemUpdate.Legend = &defaultProblem
		problemUpdate.Title = &defaultProperties.Title
		problemUpdate.TimeLimit = &defaultProperties.TimeLimit
		problemUpdate.MemoryLimit = &defaultProperties.MemoryLimit
	}
	if err := u.UpdateProblem(ctx, id, problemUpdate); err != nil {
		return err
	}

	return nil
}

func readProblem(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	problemData, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return string(problemData), nil
}

func readProperties(f *zip.File) (ProblemProperties, error) {
	rc, err := f.Open()
	if err != nil {
		return ProblemProperties{}, err
	}
	defer rc.Close()
	var properties ProblemProperties
	if err := json.NewDecoder(rc).Decode(&properties); err != nil {
		return ProblemProperties{}, err
	}
	return properties, nil
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
