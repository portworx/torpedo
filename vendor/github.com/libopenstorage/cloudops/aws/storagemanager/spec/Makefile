gen:
	go run generator.go
	sed -i '1 s/^/# This is an auto-generated file. DO NOT EDIT.\n/' aws.yaml
	cp aws.yaml ../testspecs/
	mv aws.yaml ../../../specs/decisionmatrix/
