package cmd

import (
	"fmt"
	"strconv"

	"github.com/dustin/go-humanize"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ValueCPUTotal struct {
	Total *float64
}

type ValueMemSize struct {
	Size boshdir.VMInfoVitalsMemSize
}

type ValueMemIntSize struct {
	Size boshdir.VMInfoVitalsMemIntSize
}

type ValueDiskSize struct {
	Size boshdir.VMInfoVitalsDiskSize
}

type ValueUptime struct {
	Secs *uint64
}

func NewValueStringPercent(str string) boshtbl.Value {
	return boshtbl.NewValueSuffix(boshtbl.NewValueString(str), "%")
}

func (t ValueCPUTotal) String() string {
	if t.Total != nil {
		return fmt.Sprintf("%.1f%%", *t.Total)
	}
	return ""
}

func (t ValueCPUTotal) Value() boshtbl.Value            { return t }
func (t ValueCPUTotal) Compare(other boshtbl.Value) int { panic("Never called") }

func (t ValueMemSize) String() string {
	if len(t.Size.Percent) == 0 || len(t.Size.KB) == 0 {
		return ""
	}

	kb, err := strconv.ParseUint(t.Size.KB, 10, 64)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s%% (%s)", t.Size.Percent, humanize.Bytes(kb*1000))
}

func (t ValueMemSize) Value() boshtbl.Value            { return t }
func (t ValueMemSize) Compare(other boshtbl.Value) int { panic("Never called") }

func (t ValueMemIntSize) String() string {
	if t.Size.Percent != nil && t.Size.KB != nil {
		return fmt.Sprintf("%.1f%% (%s)", *t.Size.Percent, humanize.Bytes((*t.Size.KB)*1000))
	}
	return ""
}

func (t ValueMemIntSize) Value() boshtbl.Value            { return t }
func (t ValueMemIntSize) Compare(other boshtbl.Value) int { panic("Never called") }

func (t ValueDiskSize) String() string {
	if len(t.Size.Percent) > 0 && len(t.Size.InodePercent) > 0 {
		return fmt.Sprintf("%s%% (%si%%)", t.Size.Percent, t.Size.InodePercent)
	}
	return ""
}

func (t ValueDiskSize) Value() boshtbl.Value            { return t }
func (t ValueDiskSize) Compare(other boshtbl.Value) int { panic("Never called") }

func (t ValueUptime) String() string {
	if t.Secs != nil {
		days := *t.Secs / 60 / 60 / 24
		hrs := *t.Secs / 60 / 60 % 24
		mins := *t.Secs / 60 % 60
		remSecs := *t.Secs % 60
		return fmt.Sprintf("%dd %dh %dm %ds", days, hrs, mins, remSecs)
	}
	return ""
}

func (t ValueUptime) Value() boshtbl.Value            { return t }
func (t ValueUptime) Compare(other boshtbl.Value) int { panic("Never called") }
