package args

type args struct {
	Context string
	EnvFile string

	Name      string
	Namespace string
	File      string

	AddReleaseCmd struct {
		Chart   string
		Version string
	}

	AddRepoCmd struct {
		URL string
	}

	AddSingleCmd struct {
		Create bool
	}

	UpdateCmd struct {
		List    bool
		Version string
	}

	ImportCmd struct {
		All bool
	}

	Confirm bool // confirm before executing
	Delete  bool
	Deploy  bool
	Force   bool
	DryRun  bool

	DeployCmd struct {
		All bool
	}

	Debug          bool
	NonInteractive bool
}

var Args = &args{}
