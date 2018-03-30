This controller watches Pod creations, updates and deletions.

Build:

```bash
kube-controller-gen
```

Run:
```bash
go run *.go
```

This will use `~/.kube/config` by default. Use a different config by setting the environment variable `KUBECONFIG` or the command line parameter `-kubeconfig`.
