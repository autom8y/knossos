---
name: {{.SkillName}}
description: "{{.Name}} procession reference. Use when: navigating {{.Name}} stations, understanding station goals, checking procession progress. Triggers: {{.Name}}, {{.SkillName}}, {{.Name}} stations."
---

# {{.Name}} Procession Reference

{{.Description}}

## Station Map

{{.StationTable}}

## Station Goals

{{range .Stations}}- **{{.Name}}** ({{.Rite}}{{if .AltRite}}, alt: {{.AltRite}}{{end}}): {{.Goal}}
{{end}}

## Workflow

- **Artifact directory**: `{{.ArtifactDir}}`
- **Total stations**: {{.StationCount}}
- **Entry point**: `/{{.Name}}`
- **First station**: {{.FirstStation}} ({{.FirstRite}} rite)

## Handoff Artifacts

Each station transition produces a handoff artifact at:
`{{.ArtifactDir}}/HANDOFF-{source}-to-{target}.md`

See `procession-ref` skill for the handoff schema and transition protocol.

## CLI Commands

| Command | Description |
|---------|-------------|
| `ari procession status` | Show current procession state |
| `ari procession proceed` | Advance to next station |
| `ari procession recede --to={station}` | Roll back to a previous station |
| `ari procession abandon` | Terminate the procession |
