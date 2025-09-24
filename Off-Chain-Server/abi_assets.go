package abiassets

import "embed"

// Files embeds ABI JSON files for Factory, Pool, Oracle
//go:embed abis/*.json
var Files embed.FS
