# /default

The `default` folder contains the default configuration to be used in deployment.

The `quickstart.sh` script works by making copies of the config files with `default`
to their correct locations, collecting info from the user (e.g domain, email, etc)
and then replacing the variables found in the copied default files with the values which
were provided by the user.

## Example Flow
This example flow will use the default file `/default/kratos/kratos.yml` as an example.

1. The user runs `/quickstart.sh`. It detects that `/deploy/kratos/kratos.yml`
does not exist, so it makes a copy of `/default/kratos/kratos.yml` to the location
`/deploy/kratos/kratos.yml`.
2. In the quickstart script, the user will information about their deployment
environment.
3. The information provided by the user which is relevant to `kratos.yml` will
be populated in the `/deploy/kratos/kratos.yml` (copied from default) by replacing
the placeholders (e.g `{{DOMAIN}}`) with the correct values.
4. If the user ever chooses to re-run the quickstart script, they will be given
the option to overwrite their config files. If they choose to do so, the file at
`/deploy/kratos/kratos.yml` will be moved to `/deploy/backup/kratos/kratos.yml`
and a fresh copy of `kratos.yml` will be copied from the default to the deploy location.