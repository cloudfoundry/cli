package domain

type ExperimentalVolumeMount struct {
	ContainerPath string                         `json:"container_path"`
	Mode          string                         `json:"mode"`
	Private       ExperimentalVolumeMountPrivate `json:"private"`
}

type ExperimentalVolumeMountPrivate struct {
	Driver  string `json:"driver"`
	GroupID string `json:"group_id"`
	Config  string `json:"config"`
}
