package version

var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unset"
)

func Info() (string, string, string) {
	return Version, Commit, BuildTime
}
