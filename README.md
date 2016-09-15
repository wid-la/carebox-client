This app generate the payload for Carebox.

### Usage

```
docker run --rm -v $(pwd):/home carebox-client -d deployer.prod.com -t xxXxx \
  -u registry.docker.io -l login -p pwd -e login@test.com \
  -b develop -n nimbus -x CONT_NAME:nimbus42 --insecure
```

### Help

```
  -b, --branch string             Branch name. Compose File Location
  -t, --deployerToken string      Deployer Token
  -d, --deployerUrl string        Deployer URL
  -x, --extra string              set extra vars
      --insecure                  HTTP insecure connection
  -n, --name string               Project name
  -e, --registryEmail string      Registry Email
  -l, --registryLogin string      Registry Login
  -p, --registryPassword string   Registry Password
  -u, --registryUrl string        Registry URL
```