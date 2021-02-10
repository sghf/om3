package main

import (
	"fmt"

	"opensvc.com/opensvc/core/converters"
	"opensvc.com/opensvc/core/keywords"
	"opensvc.com/opensvc/core/resource"
)

const (
	DriverGroup = "ip"
	DriverName = "route"
)

type R struct {
	Destination string	`json:"destination"`
	Gateway string		`json:"gateway"`
	Netns string		`json:"netns"`
	resource.Resource
}

func (r R) Manifest() resource.ManifestType {
	return resource.ManifestType {
		Group: DriverGroup,
		Name: DriverName,
		Keywords: []keywords.Keyword {
			keywords.Keyword {
				Name: "netns",
				Scopable: true,
				Required: true,
				Text: "the resource id of the container to plumb the ip into.",
				Example: "container#0",
			},
			keywords.Keyword {
				Name: "spec",
				Scopable: true,
				Required: true,
				Convert: converters.ConverterSHLEX,
				Text: "the resource id of the container to plumb the ip into.",
				Example: "container#0",
			},
		},
	}
}

func (r R) String() string {
        return fmt.Sprintf("<Resource %s.%s %s>", DriverGroup, DriverName, r.ResourceId)
}
