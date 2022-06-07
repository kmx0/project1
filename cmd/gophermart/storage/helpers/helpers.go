package helpers

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/docker/docker/api/types"
// 	"github.com/docker/docker/api/types/container"
// 	"github.com/docker/docker/api/types/filters"
// 	"github.com/docker/docker/client"
// )

// // Its funcs only for testing environment
// func stopDB(id string) error {
// 	cli, err := client.NewEnvClient()
// 	if err != nil {
// 		return err
// 	}

// 	err = cli.ContainerKill(context.Background(), id, "")
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func startDB() (dburl string, id string, err error) {
// 	image := "postgres"
// 	cli, err := client.NewEnvClient()
// 	if err != nil {
// 		return "", "", err
// 	}
// 	imgFilter := filters.NewArgs()
// 	imgFilter.Add("reference", image)
// 	images, err := cli.ImageList(context.Background(), types.ImageListOptions{Filters: imgFilter})
// 	if err != nil {
// 		return "", "", err
// 	}
// 	if len(images) == 0 {
// 		_, err = cli.ImagePull(context.Background(), "docker.io/library", types.ImagePullOptions{})
// 		if err != nil {
// 			return "", "", err
// 		}
// 		var count = 0
// 		fmt.Print("Pulling")
// 		for count == 0 {
// 			images, err = cli.ImageList(context.Background(), types.ImageListOptions{Filters: imgFilter})
// 			count = len(images)
// 			time.Sleep(time.Second)
// 			fmt.Print("*")
// 		}
// 		fmt.Println("Done")
// 	}
// 	health := &container.HealthConfig{
// 		Interval: time.Second,
// 		Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
// 	}
// 	containerConfig := &container.Config{
// 		Image: image,
// 		Env: []string{
// 			"PGPASSWORD='password'",
// 			"PGUSER=postgres",
// 		},
// 		Healthcheck: health,
// 	}
// 	hostConfig := &container.HostConfig{
// 		AutoRemove: true,
// 		PublishAllPorts: ,
// 	}

// }
