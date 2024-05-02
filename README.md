# Fuzz

### Compute Coverage
```bash
$ go build -covermode=atomic -cover . && rm -rf cover && mkdir cover && GOCOVERDEBUG=0 GOCOVERDIR=cover ./muskinfra
```

Hit some APIs

```bash
## GET users
$ curl localhost:4000/users
## POST user
curl -d '{"username": "ash"}' localhost:4000/user
## GET users (again)
$ curl localhost:4000/users
```

View Coverage

```bash
$ curl localhost:4000/coverage
{"count":59,"coverage":"51.30%","stmt":115}
```

The coverage output might be a little different at the time of invocation.

You can cross verify these results using the existing tools as well.

```bash
$ curl localhost:4000/exit
```

The above command will make the program exit gracefully and it should write the coverage data into a folder called `cover/` in the current working directory. You can then run the following commands to view the coverage in HTML format. Note this version will always be a little higher than the previous once since we cannot test the `/exit` endpoint as part of `/coverage`.

```bash
$ go tool covdata textfmt -i=cover -o profile.txt && go tool cover -html=profile.txt
```
