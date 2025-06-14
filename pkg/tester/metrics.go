package tester

import (
	"bufio"
	"fmt"
	"github.com/Vyacheslav1557/tester/pkg"
	"os"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	CommandBeingTimed          string        `json:"command_being_timed"`
	UserTime                   float64       `json:"user_time"`
	SystemTime                 float64       `json:"system_time"`
	CPUPercent                 int           `json:"cpu_percent"`
	ElapsedTime                time.Duration `json:"elapsed_time"`
	AverageSharedTextSize      int           `json:"average_shared_text_size"`
	AverageUnsharedDataSize    int           `json:"average_unshared_data_size"`
	AverageStackSize           int           `json:"average_stack_size"`
	AverageTotalSize           int           `json:"average_total_size"`
	MaximumResidentSetSize     int           `json:"maximum_resident_set_size"`
	AverageResidentSetSize     int           `json:"average_resident_set_size"`
	MajorPageFaults            int           `json:"major_pagefaults"`
	MinorPageFaults            int           `json:"minor_pagefaults"`
	VoluntaryContextSwitches   int           `json:"voluntary_context_switches"`
	InvoluntaryContextSwitches int           `json:"involuntary_context_switches"`
	Swaps                      int           `json:"swaps"`
	BlockInputOperations       int           `json:"block_input_operations"`
	BlockOutputOperations      int           `json:"block_output_operations"`
	MessagesSent               int           `json:"messages_sent"`
	MessagesReceived           int           `json:"messages_received"`
	SignalsDelivered           int           `json:"signals_delivered"`
	PageSize                   int           `json:"page_size"`
	ExitStatus                 int           `json:"exit_status"`
	//ElapsedTimeHours           int     `json:"elapsed_time_hours"`
	//ElapsedTimeMinutes         int     `json:"elapsed_time_minutes"`
	//ElapsedTimeSeconds         int     `json:"elapsed_time_seconds"`
	//ElapsedTimeCentiseconds    int     `json:"elapsed_time_centiseconds"`
	//ElapsedTimeTotalSeconds    float64 `json:"elapsed_time_total_seconds"`
}

func parseInt(s string) (int, error) {
	i64, err := strconv.ParseInt(s, 10, 64)
	return int(i64), err
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseMetrics(path string) (*Metrics, error) {
	const op = "parseMetrics"

	metricsFile, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "failed to open metrics file")
	}
	defer metricsFile.Close()

	metrics := &Metrics{}

	scanner := bufio.NewScanner(metricsFile)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			// Could potentially log a warning for unparseable lines
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Command being timed":
			metrics.CommandBeingTimed = strings.Trim(value, `"`) // Remove surrounding quotes
		case "User time (seconds)":
			metrics.UserTime, err = parseFloat(value)
		case "System time (seconds)":
			metrics.SystemTime, err = parseFloat(value)
		case "Percent of CPU this job got":
			// Remove the '%' sign before parsing
			value = strings.TrimSuffix(value, "%")
			metrics.CPUPercent, err = parseInt(value)
		case "Elapsed (wall clock) time (h:mm:ss or m:ss)":
			metrics.ElapsedTime, err = ParseElapsedTime(value)
		case "Average shared text size (kbytes)":
			metrics.AverageSharedTextSize, err = parseInt(value)
		case "Average unshared data size (kbytes)":
			metrics.AverageUnsharedDataSize, err = parseInt(value)
		case "Average stack size (kbytes)":
			metrics.AverageStackSize, err = parseInt(value)
		case "Average total size (kbytes)":
			metrics.AverageTotalSize, err = parseInt(value)
		case "Maximum resident set size (kbytes)":
			metrics.MaximumResidentSetSize, err = parseInt(value)
		case "Average resident set size (kbytes)":
			metrics.AverageResidentSetSize, err = parseInt(value)
		case "Major (requiring I/O) page faults":
			metrics.MajorPageFaults, err = parseInt(value)
		case "Minor (reclaiming a frame) page faults":
			metrics.MinorPageFaults, err = parseInt(value)
		case "Voluntary context switches":
			metrics.VoluntaryContextSwitches, err = parseInt(value)
		case "Involuntary context switches":
			metrics.InvoluntaryContextSwitches, err = parseInt(value)
		case "Swaps":
			metrics.Swaps, err = parseInt(value)
		case "File system inputs":
			metrics.BlockInputOperations, err = parseInt(value)
		case "File system outputs":
			metrics.BlockOutputOperations, err = parseInt(value)
		case "Socket messages sent":
			metrics.MessagesSent, err = parseInt(value)
		case "Socket messages received":
			metrics.MessagesReceived, err = parseInt(value)
		case "Signals delivered":
			metrics.SignalsDelivered, err = parseInt(value)
		case "Page size (bytes)":
			metrics.PageSize, err = parseInt(value)
		case "Exit status":
			metrics.ExitStatus, err = parseInt(value)
		default:
			return nil, pkg.Wrap(pkg.ErrInternal, fmt.Errorf("unrecognized key '%s'", key), op, "")
			// You could log unrecognized keys here if needed
		}

		// Check for parsing errors immediately after potential assignment
		if err != nil {
			return nil, pkg.Wrap(pkg.ErrInternal, fmt.Errorf("failed to parse '%s' value '%s': %w", key, value, err), op, "")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "error reading input")
	}

	return metrics, nil
}

func ParseElapsedTime(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	var duration time.Duration

	if len(parts) == 3 {
		// Формат h:mm:ss
		hours, err := time.ParseDuration(parts[0] + "h")
		if err != nil {
			return 0, err
		}
		minutes, err := time.ParseDuration(parts[1] + "m")
		if err != nil {
			return 0, err
		}
		seconds, err := time.ParseDuration(parts[2] + "s")
		if err != nil {
			return 0, err
		}
		duration = hours + minutes + seconds
	} else if len(parts) == 2 {
		// Формат m:ss
		minutes, err := time.ParseDuration(parts[0] + "m")
		if err != nil {
			return 0, err
		}
		seconds, err := time.ParseDuration(parts[1] + "s")
		if err != nil {
			return 0, err
		}
		duration = minutes + seconds
	} else {
		return 0, fmt.Errorf("invalid time format: %s", s)
	}

	return duration, nil
}
