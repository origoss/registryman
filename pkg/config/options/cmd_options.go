package options

var DryRun bool = false
var ForceDelete bool = false

func SetOptions(dryRun bool, forceDelete bool) {
	DryRun = dryRun
	ForceDelete = forceDelete
}
