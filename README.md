## URL Shortener

This is learning project for Yandex Practicum's Go Course.

Shorty can 
 - store original url and provide shortened url back
 - redirect to original url

It also provides simple management operations: listing and deleting urls for specific user, which is authorized via jwt cookie.

Storage backends are
 - in-memory
 - in-memory with fs persistence
 - database (PostgreSQL)

## Development

To be able to run all the tests you need to install
 - docker
 - [mockery](https://vektra.github.io/mockery/latest/installation/#installation)
 - [golangci-lint](https://golangci-lint.run/usage/install/)
 - `shortenertestbeta`, `random`, `statictest` -> [here](https://github.com/Yandex-Practicum/go-autotests/releases)

```bash
git clone https://github.com/adwski/shorty.git
cd shorty

go mod download

# generate mock, run lint, unit tests
make test

# run database tests and shortenertest (it will spawn docker compose project)
make integration-tests
```
