package main

import (
    "context" // Added missing import
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/tranthien1794/terraform-provider-oneprovider"
)

func main() {
    providerserver.Serve(context.Background(), func() provider.Provider {
        return oneprovider.New()
    }, providerserver.ServeOpts{
        Address: "registry.terraform.io/tranthien1794/oneprovider",
    })
}
