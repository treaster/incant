module github.com/treaster/ssg

go 1.21.1

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/itchyny/gojq v0.12.16
	github.com/stretchr/testify v1.9.0
	github.com/treaster/golist v0.0.0-00010101000000-000000000000
	github.com/treaster/shire v0.0.0-00010101000000-000000000000
	github.com/yuin/goldmark v1.7.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/treaster/shire => ../shire

replace github.com/treaster/golist => ../golist
