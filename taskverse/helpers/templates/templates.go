package templates

import "embed"

//go:embed resources/*
var TemplateFiles embed.FS
