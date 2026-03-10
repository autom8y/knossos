package resolution

import "io/fs"

// RiteChain builds the standard 5-tier rite resolution chain.
// Tiers are ordered highest-priority first: project > user > org > platform > embedded.
// Empty directory strings are automatically skipped by NewChain.
func RiteChain(projectDir, userDir, orgDir, platformDir string, embedded fs.FS) *Chain {
	tiers := []Tier{
		{Label: "project", Dir: projectDir},
		{Label: "user", Dir: userDir},
		{Label: "org", Dir: orgDir},
		{Label: "platform", Dir: platformDir},
	}
	if embedded != nil {
		tiers = append(tiers, Tier{Label: "embedded", Dir: "rites", FS: embedded})
	}
	return NewChain(tiers...)
}

// ProcessionChain builds the standard 5-tier procession resolution chain.
// Tiers are ordered highest-priority first: project > user > org > platform > embedded.
// Empty directory strings are automatically skipped by NewChain.
func ProcessionChain(projectDir, userDir, orgDir, platformDir string, embedded fs.FS) *Chain {
	tiers := []Tier{
		{Label: "project", Dir: projectDir},
		{Label: "user", Dir: userDir},
		{Label: "org", Dir: orgDir},
		{Label: "platform", Dir: platformDir},
	}
	if embedded != nil {
		tiers = append(tiers, Tier{Label: "embedded", Dir: "processions", FS: embedded})
	}
	return NewChain(tiers...)
}

// ContextChain builds the standard 4-tier context resolution chain.
// User has highest priority (user customizations override project defaults):
// user > project > org > platform.
// Empty directory strings are automatically skipped by NewChain.
func ContextChain(userDir, projectDir, orgDir, platformDir string) *Chain {
	return NewChain(
		Tier{Label: "user", Dir: userDir},
		Tier{Label: "project", Dir: projectDir},
		Tier{Label: "org", Dir: orgDir},
		Tier{Label: "platform", Dir: platformDir},
	)
}
