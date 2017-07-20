// Package statuscheck provides the API for reporting status changes from plugins to the statuscheck plugin,
// that exposes it via ETCD and HTTP.
//
// The API provides just two functions, one for registering the plugin for status change reporting and one
// for reporting status changes.
//
// To register a plugin for providing status reports, use Register() function:
//	statuscheck.Register(PluginID, probe)
//
// If probe is not nil, statuscheck will periodically probe the plugin state through the provided function,
// otherwise it is expected that the plugin itself will report state updates through ReportStateChange():
//	statuscheck.ReportStateChange(PluginID, statuscheck.OK, nil)
//
// The default status of a plugin after registering is Init.
package statuscheck
