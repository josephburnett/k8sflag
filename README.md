# k8sflag
Flag-style bindings for Kubernetes ConfigMaps.

## Rationale

There are [a lot of ways](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) to consume ConfigMap entries from a Pod--environment variables, volume mounts and direct API access.  There is also a layer of code to write which brings the configuration values into the application code where it's needed.  The `k8sflag` library provides a binding layer that brings ConfigMap values into a Golang application with a familiar, flag-style interface.

## Example

```yaml
kind: ConfigMap
data:
  name: "world"
```

```go
var name = k8sflag.String("name", "nobody")
fmt.Printf("hello %v", name.Get())
--> hello world
```

### Dynamic Option

Flags can be static or dynamically bound.  A dynamic binding will be updated when the underlying ConfigMap is updates.  This can be useful for flipping feature flags without redeploying the binary.

```yaml
kind: ConfigMap
data:
  name: "world"
```

```go
var name = k8sflag.String("name", "nobody", k8sflag.Dynamic)
fmt.Printf("hello %v", name.Get())
--> hello world
```

```yaml
kind: ConfigMap
data:
  name: "mundo"  # changed
```

```go
fmt.Printf("hello %v", name.Get())
--> hello mundo
```

### Required Option

Flags can also be required.  If a configuration value is not present, the flag will immediately panic.

```go
var name = k8sflag.String("missing-property", "", k8sflag.Required)
--> PANIC
```

## Try it out

```bash
$ kubectl create -f example/config.yaml
$ kubectl create -f example/pod.yaml
```

From a node in cluster:
```
node $ curl http://$POD_IP
--> hello world
```

Edit `config.yaml` and change `hello.name` property from `world` to `mundo`:
```
$ kubectl apply -f example/config.yaml
```

```
node $ curl http://$POD_IP
--> hello mundo
```

## Implementation

The current implementation uses a volume mount to provide configuration because volume mounts are dynamically updated and don't require a Kubernetes client.  But with a Kuberenetes client, the same interface can be implemented without volume mounting.
