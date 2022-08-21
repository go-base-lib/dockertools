package dockertools

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"io"
)

// ImagePull 通过 imagePullName 拉取镜像
func ImagePull(imageName string) *ApiTools {
	return defaultApiTools.ImagePull(imageName)
}

// ImagePullWithOption 根据选项拉取镜像
func ImagePullWithOption(imageName string, options *ImagePullOption) *ApiTools {
	return defaultApiTools.ImagePullWithOption(imageName, options)
}

// ImageGetByName 根据镜像名称和tag获取镜像信息， name格式为: REPOSITORY[:TAG]
func ImageGetByName(name string) (result []types.ImageSummary, resErr error) {
	return defaultApiTools.ImageGetByName(name)
}

// ImageList 获取镜像列表信息
func ImageList() (result []types.ImageSummary, resErr error) {
	return defaultApiTools.ImageList()
}

// ImageListWithOption 获取镜像列表信息通过参数, 结果通过 option.ResponseHandler 进行传递
func ImageListWithOption(option *ImageListOption) *ApiTools {
	return defaultApiTools.ImageListWithOption(option)
}

// ContainerCreate 容器创建
func ContainerCreate(imageName string, config *container.Config) (res container.ContainerCreateCreatedBody, err error) {
	return defaultApiTools.ContainerCreate(imageName, config)
}

// ContainerCreateWithOption 根据参数创建容器
func ContainerCreateWithOption(option *ContainerCreateOption) *ApiTools {
	return defaultApiTools.ContainerCreateWithOption(option)
}

// ContainerStart 容器启动, 根据最后一次创建的容器进行启动
func ContainerStart() *ApiTools {
	return defaultApiTools.ContainerStart()
}

// ContainerStartWithId 容器启动伴随ID
func ContainerStartWithId(id string) *ApiTools {
	return defaultApiTools.ContainerStartWithId(id)
}

// ContainerStartWithOption 根据启动选项启动容器
func ContainerStartWithOption(option *ContainerStartOption) *ApiTools {
	return defaultApiTools.ContainerStartWithOption(option)
}

// ContainerWait 等待最后一次创建成功的容器执行完成
func ContainerWait() *ApiTools {
	return defaultApiTools.ContainerWait()
}

// ContainerWaitById 根据容器ID等待容器执行完成
func ContainerWaitById(containerId string) *ApiTools {
	return defaultApiTools.ContainerWaitById(containerId)
}

// ContainerWaitWithOption  根据传入选项等待对应容器执行完成
func ContainerWaitWithOption(option *ContainerWaitOption) *ApiTools {
	return defaultApiTools.ContainerWaitWithOption(option)
}

// ContainerLogs 查看容器日志
func ContainerLogs(callback ResponseHandler[io.Reader]) *ApiTools {
	return defaultApiTools.ContainerLogs(callback)
}

// ContainerLogsWithContainerId 根据容器ID查看日志
func ContainerLogsWithContainerId(containerId string, callback ResponseHandler[io.Reader]) *ApiTools {
	return defaultApiTools.ContainerLogsWithContainerId(containerId, callback)
}

// ContainerLogsWithOption 根据选项查看容器日志
func ContainerLogsWithOption(option *ContainerLogsOption) *ApiTools {
	return defaultApiTools.ContainerLogsWithOption(option)
}

// ContainerRemove 删除容器
func ContainerRemove() *ApiTools {
	return defaultApiTools.ContainerRemove()
}

// ContainerRemoteWithContainerId 根据容器ID删除容器
func ContainerRemoteWithContainerId(containerId string) *ApiTools {
	return defaultApiTools.ContainerRemoteWithContainerId(containerId)
}

// ContainerRemoveWithOptions 根据选项删除容器
func ContainerRemoveWithOptions(option *ContainerRemoveOption) *ApiTools {
	return defaultApiTools.ContainerRemoveWithOptions(option)
}
