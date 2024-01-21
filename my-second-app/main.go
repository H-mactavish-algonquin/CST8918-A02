package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		stack := ctx.Stack()
		org := cfg.Require("org")

		stackRef, err := pulumi.NewStackReference(ctx, fmt.Sprintf("%v/my_first_app/%v", org, stack), nil)
		if err != nil {
			return err
		}

		ctx.Export("shopUrl", stackRef.GetOutput(pulumi.String("url")))
		return nil
	})
}
