package engine

import "github.com/pdfcpu/pdfcpu/pkg/api"

// OpenDesk is a local, embedded document processor. Do not let pdfcpu create
// or rewrite a shared user config.yml: concurrent requests/packages can race
// on Windows. The pinned pdfcpu version supports this stateless setting.
//
// pdfcpu v0.16 still exposes DisableConfigDir (deprecated); when updating the
// dependency, migrate each API call to ConfigurationModeStateless explicitly.
func init() { api.DisableConfigDir() }
