package flags

type backwardsCompatibilityType int

type BackwardsCompatibilityFlag struct{}

func (f *BackwardsCompatibilityFlag) Set(v string) {}

func (f *BackwardsCompatibilityFlag) String() string {
	return ""
}

func (f *BackwardsCompatibilityFlag) GetName() string {
	return ""
}

func (f *BackwardsCompatibilityFlag) GetShortName() string {
	return ""
}

func (f *BackwardsCompatibilityFlag) GetValue() interface{} {
	return backwardsCompatibilityType(1)
}

func (f *BackwardsCompatibilityFlag) Visible() bool {
	return false
}
