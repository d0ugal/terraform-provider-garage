.PHONY: build install test

build:
	go build -o terraform-provider-garage

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hoose/garage/0.1.0/linux_amd64
	cp terraform-provider-garage ~/.terraform.d/plugins/registry.terraform.io/hoose/garage/0.1.0/linux_amd64/

test:
	go test ./...

clean:
	rm -f terraform-provider-garage


