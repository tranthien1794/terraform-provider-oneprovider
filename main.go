package main

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/tranthien1794/terraform-provider-oneprovider/oneprovider" // Updated import
)

func main() {
    providerserver.Serve(context.Background(), func() provider.Provider {
        return oneprovider.New()
    }, providerserver.ServeOpts{
        Address: "registry.terraform.io/tranthien1794/oneprovider",
    })
}
