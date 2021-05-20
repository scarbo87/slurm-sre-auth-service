Deploy chart:
```
helm upgrade --install auth-service .helm/slurm-sre-auth-service --timeout 300s --atomic --debug
```