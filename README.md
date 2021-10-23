![GitHub Workflow Status](https://img.shields.io/github/workflow/status/RefractorGSCM/Refractor/Test?label=tests&style=flat-square)
![GitHub](https://img.shields.io/github/license/RefractorGSCM/Refractor?style=flat-square)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/RefractorGSCM/Refractor?style=flat-square)
![GitHub contributors](https://img.shields.io/github/contributors/RefractorGSCM/Refractor?style=flat-square)

# Refractor

Refractor is a game server community manager written in Go. It improves game server moderation by providing features such as:

- Easy installation with Docker
- Monitoring of multiple servers
- Multiple user accounts with advanced access control
- Live player lists
- Player infraction logging and lookup
- Chat message logging and monitoring
- Ban synchronization
- Live chat
- and more!

For a full list of features, check out the [documentation](https://refractor.dmas.dev).

## Technologies Used

The backend service (this repo) is written in Go. It uses [ORY Kratos](https://github.com/ory/kratos) for secure identity management.

PostgreSQL is the currently supported database, though support for other databases can be easily added.

Deployment is done using Docker with Nginx as a reverse proxy and LetsEncrypt for automatic SSL encryption which renews automatically.

The frontend application is written in Svelte and is available in the [Refractor-Svelte repo](https://github.com/RefractorGSCM/Refractor-Svelte/).

## Contributing

As Refractor is open source software, contributions are welcome!

As Refractor was just recently released, there is currently no development roadmap. That said, feature requests or pull requests are always welcome!

If you find any issues, please open an issue in this repository and include all relevant information including steps to reproduce.

A good first contribution to Refractor would be improvements or additions to Refractor's documentation! See the Development section below for a link to the documentation repository.

## Development

Development environment setup steps and other developer resources can be found in Refractor's [documentation](https://refractor.dmas.dev/).

Refractor's documentation source code is available in the [Refractor-Docs repository](https://github.com/RefractorGSCM/Refractor-Docs/). 

# License

```
Refractor is an open source game server community management application.
Copyright (C) 2021 Duncan Snider

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```
