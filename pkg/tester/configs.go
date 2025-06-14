package tester

import (
	"github.com/Vyacheslav1557/tester/internal/models"
	"time"
)

type Config interface {
	Image() string
	CompileTL() time.Duration
	CompileML() int64
	CompileCMD() []string
	ExecuteCMD() []string
}

func GetConfig(lang models.LanguageName) Config {
	switch lang {
	case models.Cpp:
		return CppConfig{}
	case models.Golang:
		return GolangConfig{}
	case models.Python:
		return PythonConfig{}
	}

	return nil
}

const (
	defaultCompileTL = time.Second * 20
	defaultCompileML = 256 * 1024 * 1024
)

type CppConfig struct{}

func (c CppConfig) Image() string {
	return "custom-golang:1.20custom-golang:1.20"
}

func (c CppConfig) CompileTL() time.Duration {
	return defaultCompileTL
}

func (c CppConfig) CompileML() int64 {
	return defaultCompileML
}

func (c CppConfig) CompileCMD() []string {
	return []string{
		"bash", "-c",
		"ln -s /code/source /code/source.cpp && g++ -o /code/solution /code/source.cpp",
	}
}

func (c CppConfig) ExecuteCMD() []string {
	return []string{
		"bash", "-c",
		"/usr/bin/time -v -o /dev/stderr bash -c 'ulimit -t 30; /code/solution'",
	}
}

type GolangConfig struct{}

func (c GolangConfig) Image() string {
	return "custom-golang:1.20"
}

func (c GolangConfig) CompileTL() time.Duration {
	return 60 * time.Second
}

func (c GolangConfig) CompileML() int64 {
	return defaultCompileML
}

func (c GolangConfig) CompileCMD() []string {
	return []string{
		"bash", "-c",
		"ln -s /code/source /code/source.go && GO111MODULE=off go build -o /code/solution /code/source.go",
	}
}

func (c GolangConfig) ExecuteCMD() []string {
	return []string{
		"bash", "-c",
		"/usr/bin/time -v -o /dev/stderr bash -c 'ulimit -t 30; /code/solution'",
	}
}

type PythonConfig struct{}

func (c PythonConfig) Image() string {
	return "custom-python:3.9"
}

func (c PythonConfig) CompileTL() time.Duration {
	return defaultCompileTL
}

func (c PythonConfig) CompileML() int64 {
	return defaultCompileML
}

func (c PythonConfig) CompileCMD() []string {
	return []string{
		"bash", "-c",
		`pypy3 -c 'import py_compile; py_compile.compile("/code/source", doraise=True)' && cp /code/source /code/solution`,
	}
}

func (c PythonConfig) ExecuteCMD() []string {
	return []string{
		"bash", "-c",
		"/usr/bin/time -v -o /dev/stderr bash -c 'ulimit -t 30; pypy3 /code/solution'",
	}
}
