package version

import (
	"fmt"
	"runtime"
	"time"
)

var (
	// Version is the main version of application
	Version = "None"
	// Revision is the commit revision of the application SC
	Revision = "000000"
	// Branch is the branch of the application SC
	Branch = "None"
	// BuildDate is the date it was build
	BuildDate = time.Now().UTC().String()
	// GoVersion is the version of go
	GoVersion = runtime.Version()
)

// Info reprensets the application information
type Info struct {
	Version   string
	Revision  string
	Branch    string
	BuildDate string
	GoVersion string
}

func (i Info) String() string {
	return fmt.Sprintf("Ladder version %s, build %s(%s)", i.Version, i.Revision, i.Branch)
}

// Get returns the application information
func Get() Info {
	return Info{
		Version:   Version,
		Revision:  Revision,
		Branch:    Branch,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
	}
}
