package main

import (
    "context" // Added missing import
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/yourusername/terraform-provider-oneprovider/oneprovider"
)

func main() {
    providerserver.Serve(context.Background(), func() provider.Provider {
        return oneprovider.New()
    }, providerserver.ServeOpts{
        Address: "registry.terraform.io/yourusername/oneprovider",
    })
}
