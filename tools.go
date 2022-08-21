package dockertools

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"time"
)

type withClientFn func(cli *client.Client) error

type ContextGet func() context.Context

type ResponseHandler[T any] func(res T, err error) error

type ApiTools struct {
	clientGet             ClientGet
	CtxGet                ContextGet
	cli                   *client.Client
	err                   error
	containerEndCreateRes container.ContainerCreateCreatedBody
}

type ApiOption[T any, F any] struct {
	Params          T
	Ctx             context.Context
	ResponseHandler ResponseHandler[F]
}

func checkOption[T any, F any](option ApiOption[T, F], api *ApiTools) {

}
func (a *ApiOption[T, F]) check(api *ApiTools) (res *ApiOption[T, F]) {
	res = a
	if a == nil {
		res = &ApiOption[T, F]{}
	}
	if res.Ctx == nil {
		res.Ctx = api.CtxGet()
	}

	return res
}

func (a *ApiOption[T, F]) resHandler(res F, err error, api *ApiTools) {
	if api.err == nil {
		api.err = err
	}
	if a.ResponseHandler == nil {
		return
	}
	api.err = a.ResponseHandler(res, err)
}

func NewApiTools(clientGet ClientGet) *ApiTools {
	return &ApiTools{
		clientGet: clientGet,
		CtxGet: func() context.Context {
			return context.Background()
		},
	}
}

func (a *ApiTools) WithClient(fn withClientFn) *ApiTools {
	if a.err != nil {
		return a
	}
	if a.cli == nil {
		if a.cli, a.err = a.clientGet(); a.err != nil {
			return a
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if _, a.err = a.cli.Ping(ctx); a.err != nil {
			return a
		}
	}
	a.err = fn(a.cli)
	return a
}

func (a *ApiTools) Error() (err error) {
	if a.cli != nil {
		_ = a.cli.Close()
		a.cli = nil
	}
	err = a.err
	a.err = nil
	return
}

// ImagePullOption 镜像拉取选项
type ImagePullOption = ApiOption[types.ImagePullOptions, io.Reader]

// ImagePull 镜像拉取
func (a *ApiTools) ImagePull(imageName string) *ApiTools {
	return a.ImagePullWithOption(imageName, nil)
}

// ImagePullWithOption 镜像拉取传入参数
func (a *ApiTools) ImagePullWithOption(imageName string, options *ImagePullOption) *ApiTools {
	options = options.check(a)
	return a.WithClient(func(cli *client.Client) error {
		r, err := cli.ImagePull(options.Ctx, imageName, options.Params)
		defer r.Close()
		options.resHandler(r, err, a)
		_, _ = io.Copy(io.Discard, r)
		return err
	})
}

// ImageListOption 镜像列表选项
type ImageListOption = ApiOption[types.ImageListOptions, []types.ImageSummary]

// ImageGetByName 获取镜像信息根据名称
func (a *ApiTools) ImageGetByName(name string) (result []types.ImageSummary, resErr error) {

	a.ImageListWithOption(&ImageListOption{
		Params: types.ImageListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", name)),
		},
		ResponseHandler: func(res []types.ImageSummary, err error) error {
			result = res
			resErr = err
			return err
		},
	})
	return
}

// ImageList 获取镜像列表
func (a *ApiTools) ImageList() (result []types.ImageSummary, resErr error) {
	a.ImageListWithOption(&ImageListOption{
		ResponseHandler: func(res []types.ImageSummary, err error) error {
			result = res
			resErr = err
			return err
		},
	})
	return
}

// ImageListWithOption 根据参数获取镜像列表
func (a *ApiTools) ImageListWithOption(option *ImageListOption) *ApiTools {
	option = option.check(a)
	return a.WithClient(func(cli *client.Client) error {
		res, err := cli.ImageList(option.Ctx, option.Params)
		option.resHandler(res, err, a)
		return a.err
	})
}

type ContainerCreateParam struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
	Platform         *specs.Platform
	ContainerName    string
}

type ContainerCreateOption = ApiOption[*ContainerCreateParam, container.ContainerCreateCreatedBody]

// ContainerCreate 容器创建
func (a *ApiTools) ContainerCreate(imageName string, config *container.Config) (res container.ContainerCreateCreatedBody, err error) {
	a.ContainerCreateWithCallback(imageName, config, func(r container.ContainerCreateCreatedBody, e error) error {
		if e != nil {
			return e
		}
		res = r
		return e
	})
	err = a.err
	return
}

// ContainerCreateWithCallback 容器创建
func (a *ApiTools) ContainerCreateWithCallback(imageName string, config *container.Config, callback ResponseHandler[container.ContainerCreateCreatedBody]) *ApiTools {
	if config == nil {
		config = &container.Config{}
	}
	createParam := &ContainerCreateParam{
		Config: config,
	}
	createParam.Config.Image = imageName

	return a.ContainerCreateWithOption(&ContainerCreateOption{
		Params:          createParam,
		ResponseHandler: callback,
	})
}

// ContainerCreateWithOption 根据参数创建容器
func (a *ApiTools) ContainerCreateWithOption(option *ContainerCreateOption) *ApiTools {
	option = option.check(a)
	return a.WithClient(func(cli *client.Client) error {
		if option.Params == nil {
			option.Params = &ContainerCreateParam{}
		}
		res, err := cli.ContainerCreate(option.Ctx, option.Params.Config, option.Params.HostConfig, option.Params.NetworkingConfig, option.Params.Platform, option.Params.ContainerName)
		option.resHandler(res, err, a)
		if a.err == nil {
			a.containerEndCreateRes = res
		}
		return a.err
	})
}

// ContainerStartParam 容器启动参数
type ContainerStartParam struct {
	Id     string
	Option types.ContainerStartOptions
}

// ContainerStartOption 容器启动选项
type ContainerStartOption = ApiOption[*ContainerStartParam, bool]

// ContainerStart 容器启动, 根据最后一次创建的容器进行启动
func (a *ApiTools) ContainerStart() *ApiTools {
	return a.ContainerStartWithOption(nil)
}

// ContainerStartWithId 容器启动伴随ID
func (a *ApiTools) ContainerStartWithId(id string) *ApiTools {
	return a.ContainerStartWithOption(&ContainerStartOption{
		Params: &ContainerStartParam{
			Id: id,
		},
	})
}

// ContainerStartWithOption 根据启动选项启动容器
func (a *ApiTools) ContainerStartWithOption(option *ContainerStartOption) *ApiTools {
	option = option.check(a)
	return a.WithClient(func(cli *client.Client) error {
		if option.Params == nil {
			option.Params = &ContainerStartParam{}
		}

		if option.Params.Id == "" {
			option.Params.Id = a.containerEndCreateRes.ID
		}

		if option.Params.Id == "" {
			return fmt.Errorf("未传容器ID")
		}

		if err := cli.ContainerStart(option.Ctx, option.Params.Id, option.Params.Option); err != nil {
			option.resHandler(false, err, a)
		} else {
			option.resHandler(true, err, a)
		}
		return a.err
	})
}

// ContainerWaitParam 容器等待参数
type ContainerWaitParam struct {
	Id        string
	Condition container.WaitCondition
}

// ContainerWaitResponse 容器等待结果
type ContainerWaitResponse struct {
	// Res 响应
	Res <-chan container.ContainerWaitOKBody
	// Err 错误
	Err <-chan error
}

// ContainerWaitOption 根据选项启动容器
type ContainerWaitOption = ApiOption[*ContainerWaitParam, *ContainerWaitResponse]

// ContainerWait 等待最后一次创建成功的容器执行完成
func (a *ApiTools) ContainerWait() *ApiTools {
	return a.ContainerWaitWithOption(nil)
}

// ContainerWaitById 根据容器ID等待容器执行完成
func (a *ApiTools) ContainerWaitById(containerId string) *ApiTools {
	return a.ContainerWaitWithOption(&ContainerWaitOption{
		Params: &ContainerWaitParam{Id: containerId},
	})
}

// ContainerWaitWithOption  根据传入选项等待对应容器执行完成
func (a *ApiTools) ContainerWaitWithOption(option *ContainerWaitOption) *ApiTools {
	option = option.check(a)
	return a.WithClient(func(cli *client.Client) error {
		if option.Params == nil {
			option.Params = &ContainerWaitParam{}
		}

		if option.Params.Id == "" {
			option.Params.Id = a.containerEndCreateRes.ID
		}

		if option.Params.Id == "" {
			return fmt.Errorf("未传容器ID")
		}

		if option.Params.Condition == "" {
			option.Params.Condition = container.WaitConditionNotRunning
		}

		statusCh, errCh := cli.ContainerWait(option.Ctx, option.Params.Id, option.Params.Condition)
		if option.ResponseHandler != nil {
			a.err = option.ResponseHandler(&ContainerWaitResponse{
				Res: statusCh,
				Err: errCh,
			}, nil)
		} else {
			select {
			case <-statusCh:
			case a.err = <-errCh:
			}
		}

		return a.err
	})
}

type ContainerLogsParam struct {
	ContainerId string
	Option      types.ContainerLogsOptions
}

type ContainerLogsOption = ApiOption[*ContainerLogsParam, io.Reader]

// ContainerLogs 查看容器日志
func (a *ApiTools) ContainerLogs(callback ResponseHandler[io.Reader]) *ApiTools {
	return a.ContainerLogsWithContainerId("", callback)
}

// ContainerLogsWithContainerId 根据容器ID查看日志
func (a *ApiTools) ContainerLogsWithContainerId(containerId string, callback ResponseHandler[io.Reader]) *ApiTools {
	return a.ContainerLogsWithOption(&ContainerLogsOption{
		Params: &ContainerLogsParam{
			ContainerId: containerId,
			Option: types.ContainerLogsOptions{
				ShowStderr: true,
				ShowStdout: true,
				Details:    true,
			},
		},
		ResponseHandler: callback,
	})
}

// ContainerLogsWithOption 根据选项查看容器日志
func (a *ApiTools) ContainerLogsWithOption(option *ContainerLogsOption) *ApiTools {
	option = option.check(a)
	return a.WithClient(func(cli *client.Client) error {
		if option.Params == nil {
			option.Params = &ContainerLogsParam{}
		}

		if option.Params.ContainerId == "" {
			option.Params.ContainerId = a.containerEndCreateRes.ID
		}

		if option.Params.ContainerId == "" {
			return fmt.Errorf("未传容器ID")
		}

		res, err := cli.ContainerLogs(option.Ctx, option.Params.ContainerId, option.Params.Option)
		if err == nil {
			defer res.Close()
		}
		option.resHandler(res, err, a)
		return a.err
	})
}

// ContainerRemoveParam 容器删除参数
type ContainerRemoveParam struct {
	ContainerId string
	Options     types.ContainerRemoveOptions
}

// ContainerRemoveOption 容器删除时的选项
type ContainerRemoveOption = ApiOption[*ContainerRemoveParam, bool]

// ContainerRemove 删除容器
func (a *ApiTools) ContainerRemove() *ApiTools {
	return a.ContainerRemoteWithContainerId("")
}

// ContainerRemoteWithContainerId 根据容器ID删除容器
func (a *ApiTools) ContainerRemoteWithContainerId(containerId string) *ApiTools {
	return a.ContainerRemoveWithOptions(&ContainerRemoveOption{
		Params: &ContainerRemoveParam{
			ContainerId: containerId,
		},
	})
}

// ContainerRemoveWithOptions 根据选项删除容器
func (a *ApiTools) ContainerRemoveWithOptions(option *ContainerRemoveOption) *ApiTools {
	option = option.check(a)
	return a.WithClient(func(cli *client.Client) error {
		if option.Params == nil {
			option.Params = &ContainerRemoveParam{}
		}

		if option.Params.ContainerId == "" {
			option.Params.ContainerId = a.containerEndCreateRes.ID
		}

		if option.Params.ContainerId == "" {
			return fmt.Errorf("未传容器ID")
		}

		if err := cli.ContainerRemove(option.Ctx, option.Params.ContainerId, option.Params.Options); err != nil {
			option.resHandler(false, err, a)
		} else {
			option.resHandler(true, err, a)
		}
		return a.err
	})
}
