package wkct

type job struct {
	ID       string
	Status   string
	Stdout   string
	Stderr   string
	Command  string
	ExitCode string
	User     string
}
