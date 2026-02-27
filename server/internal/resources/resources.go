package resources

import (
	_ "embed"
)

//go:embed plugin-runtime.Dockerfile
var PluginRuntimeDockerfile []byte
