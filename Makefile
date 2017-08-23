# This file is part of xmlsect.
#
# Copyright (C) 2017  David Gamba Rios
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

BUILD_FLAGS=-ldflags="-X github.com/DavidGamba/xmlsect/semver.BuildMetadata=`git rev-parse HEAD`"

test:
	go test ./...

debug:
	pwd; \
	echo ${GOPATH}; \
	ls **;

deps:
	go get -u github.com/DavidGamba/go-getoptions; \
	go get -u github.com/santhosh-tekuri/dom; \
	go get -u github.com/santhosh-tekuri/xpath;


doc:
	asciidoctor README.adoc

man:
	asciidoctor -b manpage xmlsect.adoc

open:
	open README.html

build:
	go build $(BUILD_FLAGS)

install:
	go install $(BUILD_FLAGS) xmlsect.go

rpm:
	rpmbuild -bb rpm.spec \
		--define '_rpmdir ./RPMS' \
		--define '_sourcedir ${PWD}' \
		--buildroot ${PWD}/buildroot
