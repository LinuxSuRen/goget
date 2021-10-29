```shell
argocd app create goget --repo https://github.com/LinuxSuRen/goget.git --path deploy \
  --dest-server https://kubernetes.default.svc --dest-namespace default
```

```shell
argocd app sync goget
```
