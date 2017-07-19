# This file is part of xmlsect.
#
# Copyright (C) 2017  David Gamba Rios
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

BUILD_FLAGS=-ldflags="-X github.com/davidgamba/xmlsect/semver.BuildMetadata=`git rev-parse HEAD`"

test:
	go test ./...

build:
	go build $(BUILD_FLAGS)

install:
	go install $(BUILD_FLAGS) xmlsect.go
