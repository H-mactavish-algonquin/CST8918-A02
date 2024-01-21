package main

import (
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get configuration values
		cfg := config.New(ctx, "")
		frontendPort := cfg.RequireFloat64("frontendPort")
		backendPort := cfg.RequireFloat64("backendPort")
		mongoPort := cfg.RequireFloat64("mongoPort")
		mongoHost := cfg.Require("mongoHost") // Note that strings are the default, so it's not `cfg.RequireStr`, just `cfg.Require`
		database := cfg.Require("database")
		nodeEnvironment := cfg.Require("nodeEnvironment")
		protocol := cfg.Require("protocol")

		// Pull the backend image
		backendImageName := "backend"
		backendImage, err := docker.NewRemoteImage(ctx, fmt.Sprintf("%v-image", backendImageName), &docker.RemoteImageArgs{
			Name: pulumi.String("pulumi/tutorial-pulumi-fundamentals-backend:latest"),
		})
		if err != nil {
			return err
		}
		ctx.Export("backendDockerImage", backendImage.Name)

		// Pull the frontend image
		frontendImageName := "frontend"
		frontendImage, err := docker.NewRemoteImage(ctx, fmt.Sprintf("%v-image", frontendImageName), &docker.RemoteImageArgs{
			Name: pulumi.String("pulumi/tutorial-pulumi-fundamentals-frontend:latest"),
		})
		if err != nil {
			return err
		}
		ctx.Export("frontendDockerImage", frontendImage.Name)

		// Pull the MongoDB image
		mongoImage, err := docker.NewRemoteImage(ctx, "mongo-image", &docker.RemoteImageArgs{
			Name: pulumi.String("pulumi/tutorial-pulumi-fundamentals-database-local:latest"),
		})
		if err != nil {
			return err
		}
		ctx.Export("mongoDockerImage", mongoImage.Name)

		// Create a Docker network
		network, err := docker.NewNetwork(ctx, "network", &docker.NetworkArgs{
			Name: pulumi.String(fmt.Sprintf("services-%v", ctx.Stack())),
		})
		if err != nil {
			return err
		}
		ctx.Export("containerNetwork", network.Name)

		// Create the MongoDB container
		mongoContainer, err := docker.NewContainer(ctx, "mongo-container", &docker.ContainerArgs{
			Name:  pulumi.String(fmt.Sprintf("mongo-%v", ctx.Stack())),
			Image: mongoImage.RepoDigest,
			Ports: &docker.ContainerPortArray{
				&docker.ContainerPortArgs{
					Internal: pulumi.Int(mongoPort),
					External: pulumi.Int(mongoPort),
				},
			},
			NetworksAdvanced: &docker.ContainerNetworksAdvancedArray{
				&docker.ContainerNetworksAdvancedArgs{
					Name: network.Name,
					Aliases: pulumi.StringArray{
						pulumi.String("mongo"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Create the backend container
		// Use _ instead of a variable name since this container isn't referenced
		_, err = docker.NewContainer(ctx, "backend-container", &docker.ContainerArgs{
			Name:  pulumi.String(fmt.Sprintf("backend-%v", ctx.Stack())),
			Image: backendImage.RepoDigest,
			Ports: &docker.ContainerPortArray{
				&docker.ContainerPortArgs{
					Internal: pulumi.Int(backendPort),
					External: pulumi.Int(backendPort),
				},
			},
			Envs: pulumi.StringArray{
				pulumi.String(fmt.Sprintf("DATABASE_HOST=%v", mongoHost)),
				pulumi.String(fmt.Sprintf("DATABASE_NAME=%v", database)),
				pulumi.String(fmt.Sprintf("NODE_ENV=%v", nodeEnvironment)),
			},
			NetworksAdvanced: &docker.ContainerNetworksAdvancedArray{
				&docker.ContainerNetworksAdvancedArgs{
					Name: network.Name,
					Aliases: pulumi.StringArray{
						pulumi.String(fmt.Sprintf("backend-%v", ctx.Stack())),
					},
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{
			mongoContainer,
		}))
		if err != nil {
			return err
		}

		// Create the frontend container
		_, err = docker.NewContainer(ctx, "frontend-container", &docker.ContainerArgs{
			Name:  pulumi.String(fmt.Sprintf("frontend-%v", ctx.Stack())),
			Image: frontendImage.RepoDigest,
			Ports: &docker.ContainerPortArray{
				&docker.ContainerPortArgs{
					Internal: pulumi.Int(frontendPort),
					External: pulumi.Int(frontendPort),
				},
			},
			Envs: pulumi.StringArray{
				pulumi.String(fmt.Sprintf("PORT=%v", frontendPort)),
				pulumi.String(fmt.Sprintf("HTTP_PROXY=backend-%v:%v", ctx.Stack(), backendPort)),
				pulumi.String(fmt.Sprintf("PROXY_PROTOCOL=%v", protocol)),
			},
			NetworksAdvanced: &docker.ContainerNetworksAdvancedArray{
				&docker.ContainerNetworksAdvancedArgs{
					Name: network.Name,
					Aliases: pulumi.StringArray{
						pulumi.String(fmt.Sprintf("frontend-%v", ctx.Stack())),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
