package tester

import (
	"bufio"
	"fmt"
	"github.com/Vyacheslav1557/tester/pkg"
	"math"
	"os"
	"strconv"
	"strings"
)

func isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func compareFloats(a, b float64, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func compareFiles(expectedPath, actualPath string, epsilon float64) error {
	const op = "compareFiles"

	expectedFile, err := os.Open(expectedPath)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "cannot open expected file")
	}
	defer expectedFile.Close()

	actualFile, err := os.Open(actualPath)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "cannot open actual file")
	}
	defer actualFile.Close()

	expectedScanner := bufio.NewScanner(expectedFile)
	actualScanner := bufio.NewScanner(actualFile)

	//defer func() {
	//	if err := expectedScanner.Err(); err != nil {
	//		res = CompareResult{PE, fmt.Sprintf("Error reading expected file: %v", err)}
	//	}
	//	if err := actualScanner.Err(); err != nil {
	//		res = CompareResult{PE, fmt.Sprintf("Error reading actual file: %v", err)}
	//	}
	//}()

	lineNumber := 0
	for {
		hasExpected := expectedScanner.Scan()
		hasActual := actualScanner.Scan()
		lineNumber++

		if !hasExpected && !hasActual {
			return nil // accepted
		}
		if hasExpected != hasActual {
			return pkg.Wrap(PresentationErr, nil, op,
				fmt.Sprintf("Different number of lines: file ended at line %d", lineNumber),
			)
		}

		expectedLine := expectedScanner.Text()
		actualLine := actualScanner.Text()

		expectedTokens := strings.Fields(expectedLine)
		actualTokens := strings.Fields(actualLine)

		if len(expectedTokens) != len(actualTokens) {
			return pkg.Wrap(PresentationErr, nil, op,
				fmt.Sprintf("Different number of tokens in line %d: expected %d, got %d",
					lineNumber, len(expectedTokens), len(actualTokens)),
			)
		}

		for j := 0; j < len(expectedTokens); j++ {
			expToken := expectedTokens[j]
			actToken := actualTokens[j]

			if isFloat(expToken) && isFloat(actToken) {
				expFloat, _ := strconv.ParseFloat(expToken, 64)
				actFloat, _ := strconv.ParseFloat(actToken, 64)

				if !compareFloats(expFloat, actFloat, epsilon) {
					return pkg.Wrap(WrongAnswerErr, nil, op,
						fmt.Sprintf("Different float values in line %d, position %d: expected %s, got %s",
							lineNumber, j+1, expToken, actToken),
					)
				}
			} else if expToken != actToken {
				return pkg.Wrap(WrongAnswerErr, nil, op,
					fmt.Sprintf("Different values in line %d, position %d: expected %s, got %s",
						lineNumber, j+1, expToken, actToken),
				)
			}
		}
	}
}
