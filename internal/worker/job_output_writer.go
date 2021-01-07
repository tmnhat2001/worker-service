package worker

import "strings"

type jobOutputWriter struct {
	result     strings.Builder
	outputType string
	jobID      string
	store      JobStore
}

func (w *jobOutputWriter) Write(p []byte) (int, error) {
	n, err := w.result.Write(p)
	if err != nil {
		return n, err
	}

	values := make(map[string]string)
	if w.outputType == "stdout" {
		values["Stdout"] = w.result.String()
	} else {
		values["Stderr"] = w.result.String()
	}
	w.store.UpdateJob(w.jobID, values)

	return n, nil
}
