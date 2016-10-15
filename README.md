## HEAVILY UNDER CONSTRUCTION

Presently, you are using this at your own risk.

There are definitely some improvements forth coming:

- [ ] godoc all the things
- [ ] refactor the client method, the interface will not change though 
- [ ] work though adding all of the pivnet methods (this will take awhile)


### Running the tests

Install the ginkgo executable with:

```
go get -u github.com/onsi/ginkgo/ginkgo
```

The tests require a valid Pivotal Network API token and host.

Refer to the
[official docs](https://network.pivotal.io/docs/api#how-to-authenticate)
for more details on obtaining a Pivotal Network API token.

It is advised to run the acceptance tests against the Pivotal Network integration
environment endpoint i.e. `HOST='https://pivnet-integration.cfapps.io'`.

Run the tests with the following command:

```
API_TOKEN=my-token \
HOST='https://pivnet-integration.cfapps.io' \
./bin/test_all
```

### Contributing

Please make all pull requests to the `develop` branch, and
[ensure the tests pass locally](https://github.com/pivotal-cf/go-pivnet#running-the-tests).

### Project management

The CI for this project can be found at https://sunrise.ci.cf-app.com and the
scripts can be found in the
[pivnet-resource-ci repo](https://github.com/pivotal-cf/pivnet-resource-ci).

The roadmap is captured in [Pivotal Tracker](https://www.pivotaltracker.com/projects/1474244).
