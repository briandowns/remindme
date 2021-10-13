<p align="center">
  <a href="remindme"><img src="https://images.squarespace-cdn.com/content/v1/53b65d6ee4b036664113dd10/1407066986281-VOYUV67EAU87M10D5LSZ/image-asset.jpeg" width="200" height="200" border="0" alt="remindme"></a>
</p>
<p align="center">
  <a href="https://godoc.org/github.com/briandowns/remindme"><img src="https://godoc.org/github.com/briandowns/remindme?status.svg" alt="GoDoc"></a>
  <a href="https://opensource.org/licenses/BSD-3-Clause"><img src="https://img.shields.io/badge/License-BSD%203--Clause-orange.svg?" alt="License"></a>
  <a href="https://github.com/briandowns/remindme/releases"><img src="https://img.shields.io/badge/version-0.1.0-green.svg?" alt="Version"></a>
</p>

# remindme

`remindme` is a simple application to set reminders from the CLI that integrates with your system's notification system.

## Examples

Once the server is running, `remindme -s &`, you can schedule reminders with the commands below.

```sh
remindme at 2:00 "go to the grocery store"
remindme in 10m "join the call"
```

To stop the server
```sh
$ kill -15 $(cat /tmp/remindme.pid)
```

The server startup is idempotent. If a new version is compiled or installed, just run `remindme -s &` again to start a new process
with the latest binary version.

## Building

There are 2 ways to build `remindme`.

1. Run `make`.
2. Run `docker build <CONTAINER_REPO_NAME>/remindme:v1.0.0 .`

## Contributions

* File Issue with details of the problem, feature request, etc.
* Submit a pull request and include details of what problem or feature the code is solving or implementing.

## License

`remindme` source code is available under the BSD 3 clause [License](/LICENSE).

## Contact

[@bdowns328](http://twitter.com/bdowns328)

