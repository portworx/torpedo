## Deploying consul-bench workload to the Postgresql PDS cluster

Depends on `consul-bench` and `consul-agent` images to exist in the repository.
After making changes in the `consul-*` subdirectories 
adjust the version in the Makefile and release new images:

```bash
make build
make push
```

Requires `kubectl` to be connected to the corresponding target cluster and
environment variables:

- `PDS_CLUSTER` - the name of the cluster as you see in the PDS UI
- `PDS_NS` - the namespace in target cluster (default `dev`)

Usage:

Set the cluster name:

```
export PDS_CLUSTER=<cluster_name>
```

Deploy:

```bash
./deploy.sh
```

Undeploy:

```bash
./undeploy.sh
```

You can adjust the consul-bench input parameters in the `bench.yml` file.
