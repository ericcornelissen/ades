// Copyright (C) 2023  Eric Cornelissen
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// The ades command can be used to Scan for Dangerous Expression in Actions
// (sdea -> ades) workflows and manifests - Actions being GitHub's continuous
// integrations platform.
//
// It is primarily intended to be used as a CLI application, but also exports
// its functionality for programmatic use. For programmatic use, note that this
// project does not use semantic versioning.
package main
