package types

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var faintColor = color.New(color.Faint)

type Selectable interface {
	Label() string
	Selected() string
	Details() string
	Match(input string) bool
}

type BasicSelectable struct {
	LabelStr    string
	SelectedStr string
	DetailsStr  string
}

func (b BasicSelectable) Label() string {
	return b.LabelStr
}

func (b BasicSelectable) Selected() string {
	return b.SelectedStr
}

func (b BasicSelectable) Details() string {
	return b.DetailsStr
}

func (b BasicSelectable) Match(input string) bool {
	return strings.Contains(strings.ToLower(b.LabelStr), strings.ToLower(b.LabelStr))
}

func (m ManifestRepo) Label() string {
	return color.CyanString(m.Name)
}

func (m ManifestRepo) Selected() string {
	return m.Name
}

func (m ManifestRepo) Details() string {
	return fmt.Sprintf(`
--------- Repo ----------
  %s	%s
  %s	%s
`,
		faintColor.Sprint("Name:"), m.Name,
		faintColor.Sprint("URL:"), m.URL,
	)
}

func (m ManifestRepo) Match(input string) bool {
	return strings.Contains(strings.ToLower(m.Name), strings.ToLower(input))
}

func (m ManifestRelease) Label() string {
	return fmt.Sprintf("%s - %s", color.CyanString(m.Name), color.RedString(m.Namespace))
}

func (m ManifestRelease) Selected() string {
	return m.Name
}

func (m ManifestRelease) Details() string {
	return fmt.Sprintf(`
--------- Release ----------
  %s	%s
  %s	%s
  %s	%s
  %s	%s
`,
		faintColor.Sprint("Name:"), m.Name,
		faintColor.Sprint("Namespace:"), m.Namespace,
		faintColor.Sprint("Version:"), m.Chart.Version,
		faintColor.Sprint("Chart:"), m.Chart.RepoName(),
	)
}

func (m ManifestRelease) Match(input string) bool {
	return strings.Contains(strings.ToLower(m.Name), strings.ToLower(input)) ||
		strings.Contains(strings.ToLower(m.Namespace), strings.ToLower(input)) ||
		m.Chart.Match(input)
}

func (m ManifestChart) Label() string {
	return fmt.Sprintf("%s (%s)", color.CyanString(m.RepoName()), color.RedString(m.Version))
}

func (m ManifestChart) Selected() string {
	return fmt.Sprintf("%s (%s)", m.RepoName(), m.Version)
}

func (m ManifestChart) Details() string {
	return fmt.Sprintf(`
--------- Chart ----------
  %s	%s
  %s	%s
  %s	%s
  %s	%s
`,
		faintColor.Sprint("Name:"), m.Name,
		faintColor.Sprint("Repo:"), m.Repo,
		faintColor.Sprint("Version:"), m.Version,
		faintColor.Sprint("AppVersion:"), m.AppVersion,
	)
}

func (m ManifestChart) Match(input string) bool {
	inputs := strings.Split(strings.ToLower(input), " ")
	for _, i := range inputs {
		if !(strings.Contains(strings.ToLower(m.Name), i) ||
			strings.Contains(strings.ToLower(m.Repo), i) ||
			strings.Contains(strings.ToLower(m.Version), i) ||
			strings.Contains(strings.ToLower(m.AppVersion), i)) {
			return false
		}
	}

	return true
}

func (m ManifestSingle) Label() string {
	return m.Name
}

func (m ManifestSingle) Selected() string {
	return m.Name
}

func (m ManifestSingle) Details() string {
	return fmt.Sprintf(`
--------- Single ----------
  %s	%s
%s	%t
  %s	%s
`,
		faintColor.Sprint("Name:"), m.Name,
		faintColor.Sprint("UseCreate:"), m.UseCreate,
		faintColor.Sprint("Namespace:"), m.Namespace,
	)
}

func (m ManifestSingle) Match(input string) bool {
	inputs := strings.Split(strings.ToLower(input), " ")
	for _, i := range inputs {
		if !(strings.Contains(strings.ToLower(m.Name), i) ||
			strings.Contains(strings.ToLower(m.Namespace), i)) {
			return false
		}
	}

	return true
}

func (m HelmChart) Label() string {
	return fmt.Sprintf("%s (%s)", color.CyanString(m.RepoName), color.RedString(m.Version))
}

func (m HelmChart) Selected() string {
	return fmt.Sprintf("%s (%s)", m.RepoName, m.Version)
}

func (m HelmChart) Details() string {
	return fmt.Sprintf(`
--------- Chart ----------
  %s	%s
  %s	%s
  %s	%s
  %s	%s
  %s	%s
`,
		faintColor.Sprint("Name:"), m.Name(),
		faintColor.Sprint("Repo:"), m.Repo(),
		faintColor.Sprint("Version:"), m.Version,
		faintColor.Sprint("AppVersion:"), m.AppVersion,
		faintColor.Sprint("Description:"), m.Description,
	)
}

func (m HelmChart) Match(input string) bool {
	inputs := strings.Split(strings.ToLower(input), " ")
	for _, i := range inputs {
		if !(strings.Contains(strings.ToLower(m.Name()), i) ||
			strings.Contains(strings.ToLower(m.Repo()), i) ||
			strings.Contains(strings.ToLower(m.Version), i) ||
			strings.Contains(strings.ToLower(m.AppVersion), i) ||
			strings.Contains(strings.ToLower(m.Description), i)) {
			return false
		}
	}
	return true
}

func (m HelmChartMulti) Selected() string {
	return m.RepoName
}

func (m HelmRepo) Label() string {
	return color.CyanString(m.Name)
}

func (m HelmRepo) Selected() string {
	return m.Name
}

func (m HelmRepo) Details() string {
	return fmt.Sprintf(`
--------- Repo ----------
  %s	%s
  %s	%s
`,
		faintColor.Sprint("Name:"), m.Name,
		faintColor.Sprint("URL:"), m.URL,
	)
}

func (m HelmRepo) Match(input string) bool {
	return strings.Contains(strings.ToLower(m.Name), strings.ToLower(input))
}

func (m HelmChartMultiVersion) Label() string {
	return fmt.Sprintf("%s (%s)", color.CyanString(m.Version), color.RedString(m.AppVersion))
}

func (m HelmChartMultiVersion) Selected() string {
	return m.Version
}

func (m HelmChartMultiVersion) Details() string {
	return HelmChart(m).Details()
}

func (m HelmChartMultiVersion) Match(input string) bool {
	return HelmChart(m).Match(input)
}

type SelectableString string

func (s SelectableString) String() string {
	return string(s)
}

func (s SelectableString) Label() string {
	return color.CyanString(string(s))
}

func (s SelectableString) Selected() string {
	return string(s)
}

func (s SelectableString) Details() string {
	return ""
}

func (s SelectableString) Match(input string) bool {
	inputs := strings.Split(strings.ToLower(input), " ")
	check := strings.ToLower(string(s))
	for _, i := range inputs {
		if !strings.Contains(check, i) {
			return false
		}
	}

	return true
}

func (m HelmRelease) Label() string {
	return fmt.Sprintf("%s/%s - %s", color.YellowString(m.Namespace), color.CyanString(m.Name), color.RedString(m.Version()))
}

func (m HelmRelease) Selected() string {
	return fmt.Sprintf("%s/%s", m.Namespace, m.Name)
}

func (m HelmRelease) Details() string {
	return fmt.Sprintf(`
--------- Release ----------
  %s	%s
  %s	%s
  %s	%s
  %s	%s
`,
		faintColor.Sprint("Name:"), m.Name,
		faintColor.Sprint("Namespace:"), m.Namespace,
		faintColor.Sprint("Version:"), m.Version(),
		faintColor.Sprint("Chart:"), m.Chart(),
	)
}

func (m HelmRelease) Match(input string) bool {
	return strings.Contains(strings.ToLower(m.Name), strings.ToLower(input)) ||
		strings.Contains(strings.ToLower(m.Namespace), strings.ToLower(input)) ||
		strings.Contains(strings.ToLower(m.Version()), strings.ToLower(input)) ||
		strings.Contains(strings.ToLower(m.Chart()), strings.ToLower(input))
}
