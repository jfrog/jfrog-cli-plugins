module github.com/jfrog/release-bundle-generator

go 1.14

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jfrog/jfrog-cli-core v0.0.1
	github.com/jfrog/jfrog-client-go v0.14.0
	github.com/mitchellh/copystructure v1.0.0 // indirect
	k8s.io/apimachinery v0.19.2 // indirect
	k8s.io/helm v2.16.12+incompatible
)

replace github.com/jfrog/jfrog-cli-core => github.com/jfrog/jfrog-cli-core v0.1.1-0.20200924072840-fb1c693ea103

replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v0.14.1-0.20200924070338-796778dbbcde
