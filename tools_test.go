package dockertools

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

const (
	imagePullName = "docker.io/library/alpine"

	imageName = "alpine"
	imageTag  = "latest"
)

var (
	imageStartCmd = []string{"echo", "hello world"}
)

func TestImagePull(t *testing.T) {
	a := assert.New(t)
	InitLocal()

	a.NoError(ImagePull(imagePullName).
		ImageListWithOption(&ImageListOption{
			Params: types.ImageListOptions{Filters: filters.NewArgs(filters.Arg("reference", imageName))},
			ResponseHandler: func(res []types.ImageSummary, err error) error {
				if err != nil {
					return err
				}
				if len(res) != 1 {
					return fmt.Errorf("获取镜像数据的数量不匹配")
				}

				if res[0].RepoTags[0] != fmt.Sprintf("%s:%s", imageName, imageTag) {
					return fmt.Errorf("镜像名称不匹配")
				}

				return nil
			},
		}).
		ContainerCreateWithCallback(imageName, &container.Config{Cmd: imageStartCmd}, nil).
		ContainerStart().
		ContainerLogs(func(res io.Reader, err error) error {
			if err != nil {
				return err
			}

			_, _ = io.Copy(os.Stdout, res)
			return nil
		}).
		ContainerRemove().
		Error())

}
